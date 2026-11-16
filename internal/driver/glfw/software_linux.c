//go:build linux && !android && !test_web_driver

#include "software_linux.h"

#include <errno.h>
#include <fcntl.h>
#include <limits.h>
#include <stdlib.h>
#include <string.h>

#if defined(GRANIO_X11)
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#endif

#if defined(GRANIO_WAYLAND)
#include <sys/mman.h>
#include <unistd.h>
#include <wayland-client.h>

typedef struct {
    struct wl_buffer* buffer;
    unsigned char* pixels;
    size_t length;
    int width, height;
    int busy;
} SoftwareBuffer;
#endif

struct GranioSoftwareSurface {
#if defined(GRANIO_X11)
    Display* display;
    XImage* image;
    GC gc;
#endif
#if defined(GRANIO_WAYLAND)
    struct wl_shm* shm;
    SoftwareBuffer buffers[2];
#endif
};

static void copyXRGB(unsigned char* dest, int destStride, const unsigned char* source,
                     int sourceStride, int width, int height)
{
    for (int y = 0; y < height; y++)
    {
        for (int x = 0; x < width * 4; x += 4)
        {
            dest[x] = source[x + 2];
            dest[x + 1] = source[x + 1];
            dest[x + 2] = source[x];
            dest[x + 3] = 0xff;
        }
        dest += destStride;
        source += sourceStride;
    }
}

#if defined(GRANIO_X11)
static unsigned long component(unsigned char value, unsigned long mask)
{
    const unsigned long unit = mask & (0ul - mask);
    return (((uint64_t) value * (mask / unit) + 127) / 255) * unit;
}

int granioPresentSoftwareX11(GranioSoftwareSurface* state, void* connection, uintptr_t window,
                             const unsigned char* pixels, int width, int height, int stride)
{
    Display* display = connection;
    state->display = display;
    if (!state->image || width != state->image->width || height != state->image->height)
    {
        XWindowAttributes attributes;
        if (!XGetWindowAttributes(display, window, &attributes))
            return GRANIO_SOFTWARE_FAILED;
        if (attributes.visual->class != TrueColor)
            return GRANIO_SOFTWARE_UNSUPPORTED;

        XImage* image = XCreateImage(display, attributes.visual, attributes.depth,
                                     ZPixmap, 0, NULL, width, height, 32, 0);
        if (!image)
            return GRANIO_SOFTWARE_FAILED;
        image->data = malloc((size_t) image->bytes_per_line * height);
        if (!image->data)
        {
            XDestroyImage(image);
            return GRANIO_SOFTWARE_FAILED;
        }
        if (state->image)
            XDestroyImage(state->image);
        state->image = image;
    }

    if (!state->gc)
    {
        XGCValues values = {0};
        values.graphics_exposures = False;
        state->gc = XCreateGC(display, window, GCGraphicsExposures, &values);
        if (!state->gc)
            return GRANIO_SOFTWARE_FAILED;
    }

    XImage* image = state->image;
    if (image->bits_per_pixel == 32 && image->byte_order == LSBFirst &&
        image->red_mask == 0xff0000 && image->green_mask == 0xff00 && image->blue_mask == 0xff)
    {
        copyXRGB((unsigned char*) image->data, image->bytes_per_line, pixels, stride, width, height);
    }
    else
    {
        if (!image->red_mask || !image->green_mask || !image->blue_mask)
            return GRANIO_SOFTWARE_UNSUPPORTED;
        for (int y = 0; y < height; y++)
        {
            const unsigned char* src = pixels + (size_t) y * stride;
            for (int x = 0; x < width; x++, src += 4)
            {
                const unsigned long value = component(src[0], image->red_mask) |
                                            component(src[1], image->green_mask) |
                                            component(src[2], image->blue_mask);
                XPutPixel(image, x, y, value);
            }
        }
    }

    XPutImage(display, window, state->gc, image, 0, 0, 0, 0, width, height);
    XFlush(display);
    return GRANIO_SOFTWARE_PRESENTED;
}
#endif

#if defined(GRANIO_WAYLAND)
static void registryGlobal(void* data, struct wl_registry* registry,
                           uint32_t id, const char* interface, uint32_t version)
{
    (void) version;
    struct wl_shm** shm = data;
    if (!*shm && strcmp(interface, wl_shm_interface.name) == 0)
        *shm = wl_registry_bind(registry, id, &wl_shm_interface, 1);
}

static void registryRemove(void* data, struct wl_registry* registry, uint32_t id)
{
    (void) data;
    (void) registry;
    (void) id;
}

static const struct wl_registry_listener softwareRegistryListener = {
    registryGlobal, registryRemove
};

static struct wl_shm* bindShm(struct wl_display* display)
{
    struct wl_shm* shm = NULL;
    struct wl_event_queue* queue = wl_display_create_queue(display);
    if (!queue)
        return NULL;
    // Use a private queue for discovery so no pending GLFW callbacks run here.
    struct wl_display* wrapper = wl_proxy_create_wrapper(display);
    if (wrapper)
    {
        wl_proxy_set_queue((struct wl_proxy*) wrapper, queue);
        struct wl_registry* registry = wl_display_get_registry(wrapper);
        wl_proxy_wrapper_destroy(wrapper);
        if (registry)
        {
            if (wl_registry_add_listener(registry, &softwareRegistryListener, &shm) == 0 &&
                wl_display_roundtrip_queue(display, queue) < 0 && shm)
            {
                wl_shm_destroy(shm);
                shm = NULL;
            }
            wl_registry_destroy(registry);
        }
    }
    if (shm)
        wl_proxy_set_queue((struct wl_proxy*) shm, NULL);
    wl_event_queue_destroy(queue);
    return shm;
}

static int createBackingFile(int length)
{
    int fd = memfd_create("granio-software", MFD_CLOEXEC | MFD_ALLOW_SEALING);
    if (fd >= 0)
        (void) fcntl(fd, F_ADD_SEALS, F_SEAL_SHRINK | F_SEAL_SEAL);
    else
    {
        const char* runtimeDir = getenv("XDG_RUNTIME_DIR");
        if (!runtimeDir)
            return -1;
        static const char suffix[] = "/granio-software-XXXXXX";
        const size_t base = strlen(runtimeDir);
        char* path = malloc(base + sizeof(suffix));
        if (!path)
            return -1;
        memcpy(path, runtimeDir, base);
        memcpy(path + base, suffix, sizeof(suffix));
        fd = mkostemp(path, O_CLOEXEC);
        if (fd >= 0 && unlink(path) < 0)
        {
            close(fd);
            fd = -1;
        }
        free(path);
        if (fd < 0)
            return -1;
    }
    if (posix_fallocate(fd, 0, length) != 0)
    {
        close(fd);
        return -1;
    }
    return fd;
}

static void releaseBuffer(void* data, struct wl_buffer* buffer)
{
    (void) buffer;
    ((SoftwareBuffer*) data)->busy = 0;
}

static const struct wl_buffer_listener softwareBufferListener = { releaseBuffer };

static void destroyBuffer(SoftwareBuffer* slot)
{
    if (slot->buffer)
        wl_buffer_destroy(slot->buffer);
    if (slot->length)
        munmap(slot->pixels, slot->length);
    *slot = (SoftwareBuffer) {0};
}

static int createBuffer(struct wl_shm* shm, SoftwareBuffer* slot, int width, int height)
{
    const int length = width * height * 4;
    const int fd = createBackingFile(length);
    if (fd < 0)
        return 0;

    void* pixels = mmap(NULL, length, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    struct wl_shm_pool* pool = NULL;
    if (pixels != MAP_FAILED)
        pool = wl_shm_create_pool(shm, fd, length);
    close(fd);

    struct wl_buffer* buffer = NULL;
    if (pool)
    {
        buffer = wl_shm_pool_create_buffer(pool, 0, width, height, width * 4, WL_SHM_FORMAT_XRGB8888);
        wl_shm_pool_destroy(pool);
    }
    if (!buffer || wl_buffer_add_listener(buffer, &softwareBufferListener, slot) < 0)
    {
        if (buffer)
            wl_buffer_destroy(buffer);
        if (pixels != MAP_FAILED)
            munmap(pixels, length);
        return 0;
    }
    *slot = (SoftwareBuffer) { buffer, pixels, length, width, height, 0 };
    return 1;
}

int granioPresentSoftwareWayland(GranioSoftwareSurface* state, void* connection, void* target,
                                 const unsigned char* pixels, int width, int height, int stride,
                                 int surfaceWidth, int surfaceHeight)
{
    struct wl_display* display = connection;
    struct wl_surface* surface = target;
    if (!state->shm)
    {
        state->shm = bindShm(display);
        if (!state->shm)
            return GRANIO_SOFTWARE_UNSUPPORTED;
    }

    SoftwareBuffer* slot = NULL;
    for (size_t i = 0; i < sizeof(state->buffers) / sizeof(state->buffers[0]); i++)
    {
        SoftwareBuffer* candidate = &state->buffers[i];
        // Retired buffers have separate backing files. The compositor can still read them.
        if (candidate->buffer && (candidate->width != width || candidate->height != height))
            destroyBuffer(candidate);
        if (!candidate->busy && (!slot || candidate->buffer))
            slot = candidate;
    }
    if (!slot)
        return GRANIO_SOFTWARE_DEFERRED;
    if (!slot->buffer && !createBuffer(state->shm, slot, width, height))
        return GRANIO_SOFTWARE_FAILED;

    copyXRGB(slot->pixels, width * 4, pixels, stride, width, height);
    slot->busy = 1;
    wl_surface_attach(surface, slot->buffer, 0, 0);
    if (wl_surface_get_version(surface) >= WL_SURFACE_DAMAGE_BUFFER_SINCE_VERSION)
        wl_surface_damage_buffer(surface, 0, 0, width, height);
    else
        wl_surface_damage(surface, 0, 0, surfaceWidth, surfaceHeight);
    wl_surface_commit(surface);
    if (wl_display_flush(display) < 0 && errno != EAGAIN)
        return GRANIO_SOFTWARE_FAILED;
    return GRANIO_SOFTWARE_PRESENTED;
}
#endif

GranioSoftwareSurface* granioCreateSoftwareSurface(void)
{
    return calloc(1, sizeof(GranioSoftwareSurface));
}

void granioDestroySoftwareSurface(GranioSoftwareSurface* state)
{
    if (!state)
        return;
#if defined(GRANIO_X11)
    if (state->image)
        XDestroyImage(state->image);
    if (state->gc)
        XFreeGC(state->display, state->gc);
#endif
#if defined(GRANIO_WAYLAND)
    for (size_t i = 0; i < sizeof(state->buffers) / sizeof(state->buffers[0]); i++)
        destroyBuffer(&state->buffers[i]);
    if (state->shm)
        wl_shm_destroy(state->shm);
#endif
    free(state);
}
