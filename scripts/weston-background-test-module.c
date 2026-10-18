/* Owned Weston test fixture. Implements protocol validation/capability events,
 * not a blur renderer. Compositor-side first-buffer inspection is independent
 * of Fyne's Capture. All readback is confined to this test module. */
#define _GNU_SOURCE
#include <libweston/libweston.h>
#include <wayland-server-core.h>
#include "ext-background-effect-v1-server-protocol.h"
#include "fractional-scale-v1-server-protocol.h"
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct bound_manager { struct wl_list link; struct wl_resource *resource; };
struct surface_extension {
    struct wl_list link;
    struct wl_resource *resource;
    struct weston_surface *surface;
    struct wl_listener destroy;
    int fractional;
};
struct observed_surface {
    struct weston_surface *surface;
    struct wl_listener commit, destroy;
    int frames, last_alpha;
};
static struct weston_compositor *compositor;
static struct wl_global *background_global, *fractional_global;
static struct wl_list managers, extensions;
static uint32_t capabilities = 1, preferred_scale = 120;
static struct wl_listener created;

static void destroy_request(struct wl_client *client, struct wl_resource *resource) { wl_resource_destroy(resource); }
static void manager_destroyed(struct wl_resource *resource) {
    struct bound_manager *m = wl_resource_get_user_data(resource);
    wl_list_remove(&m->link); free(m);
}
static void surface_destroyed(struct wl_listener *listener, void *data) {
    struct surface_extension *e = wl_container_of(listener, e, destroy);
    wl_list_remove(&e->destroy.link); wl_list_init(&e->destroy.link); e->surface = NULL;
}
static void extension_destroyed(struct wl_resource *resource) {
    struct surface_extension *e = wl_resource_get_user_data(resource);
    fprintf(stderr, "extension destroy surface=%u fractional=%d\n", e->surface && e->surface->resource ? wl_resource_get_id(e->surface->resource) : 0, e->fractional);
    wl_list_remove(&e->destroy.link); wl_list_remove(&e->link); free(e);
}
static struct surface_extension *new_extension(struct wl_client *client, uint32_t id,
        struct wl_resource *surface_resource, const struct wl_interface *interface, int fractional) {
    struct weston_surface *surface = wl_resource_get_user_data(surface_resource);
    struct surface_extension *e;
    wl_list_for_each(e, &extensions, link) {
        if (e->surface == surface && e->fractional == fractional) {
            wl_client_post_implementation_error(client, "duplicate surface extension"); return NULL;
        }
    }
    e = calloc(1, sizeof(*e));
    if (!e) { wl_client_post_no_memory(client); return NULL; }
    e->resource = wl_resource_create(client, interface, 1, id);
    if (!e->resource) { free(e); wl_client_post_no_memory(client); return NULL; }
    e->surface = surface; e->fractional = fractional;
    e->destroy.notify = surface_destroyed;
    wl_signal_add(&surface->destroy_signal, &e->destroy);
    wl_list_insert(&extensions, &e->link);
    fprintf(stderr,"extension create surface=%u fractional=%d\n",wl_resource_get_id(surface_resource),fractional);
    return e;
}
static void blur_region(struct wl_client *client, struct wl_resource *resource, struct wl_resource *region) {
    struct surface_extension *e = wl_resource_get_user_data(resource);
    if (!e->surface) { wl_resource_post_error(resource, EXT_BACKGROUND_EFFECT_SURFACE_V1_ERROR_SURFACE_DESTROYED,"surface destroyed"); return; }
    fprintf(stderr,"blur region surface=%u enabled=%d capability=%u\n",wl_resource_get_id(e->surface->resource),region!=NULL,capabilities);
}
static const struct ext_background_effect_surface_v1_interface effect_impl = {destroy_request, blur_region};
static void get_effect(struct wl_client *client, struct wl_resource *manager, uint32_t id, struct wl_resource *surface) {
    struct surface_extension *e = new_extension(client,id,surface,&ext_background_effect_surface_v1_interface,0);
    if (e) wl_resource_set_implementation(e->resource,&effect_impl,e,extension_destroyed);
}
static const struct ext_background_effect_manager_v1_interface manager_impl = {destroy_request,get_effect};
static void bind_background(struct wl_client *client, void *data, uint32_t version, uint32_t id) {
    struct bound_manager *m = calloc(1,sizeof(*m));
    if (!m) {wl_client_post_no_memory(client);return;}
    m->resource = wl_resource_create(client,&ext_background_effect_manager_v1_interface,1,id);
    if (!m->resource) {free(m);wl_client_post_no_memory(client);return;}
    wl_list_insert(&managers,&m->link);
    wl_resource_set_implementation(m->resource,&manager_impl,m,manager_destroyed);
    ext_background_effect_manager_v1_send_capabilities(m->resource,capabilities);
}
static const struct wp_fractional_scale_v1_interface fractional_impl = {destroy_request};
static void get_fractional(struct wl_client *client,struct wl_resource *manager,uint32_t id,struct wl_resource *surface) {
    struct surface_extension *e=new_extension(client,id,surface,&wp_fractional_scale_v1_interface,1);
    if(e) {wl_resource_set_implementation(e->resource,&fractional_impl,e,extension_destroyed);wp_fractional_scale_v1_send_preferred_scale(e->resource,preferred_scale);}
}
static const struct wp_fractional_scale_manager_v1_interface fractional_manager_impl={destroy_request,get_fractional};
static void bind_fractional(struct wl_client *client,void *data,uint32_t version,uint32_t id) {
    struct wl_resource *r=wl_resource_create(client,&wp_fractional_scale_manager_v1_interface,1,id);
    if(r)wl_resource_set_implementation(r,&fractional_manager_impl,NULL,NULL); else wl_client_post_no_memory(client);
}
static void observed_destroy(struct wl_listener *listener,void *data) {
    struct observed_surface *s=wl_container_of(listener,s,destroy);
    wl_list_remove(&s->commit.link);wl_list_remove(&s->destroy.link);free(s);
}
static void observed_commit(struct wl_listener *listener,void *data) {
    struct observed_surface *s=wl_container_of(listener,s,commit);
    struct weston_surface *surface=s->surface;
    if(!surface->buffer_ref.buffer || surface->width<300 || surface->height<200)return;
    int width,height; weston_surface_get_content_size(surface,&width,&height);
    if(width<=0 || height<=0 || width>4096 || height>4096)return;
    size_t size=(size_t)width*height*4;
    uint32_t *pixels=malloc(size);
    if(!pixels)return;
    if(weston_surface_copy_content(surface,pixels,size,0,0,width,height)<0){free(pixels);return;}
    int zeros=0,nonopaque=0,invalid=0;
    for(size_t i=0;i<(size_t)width*height;i++) {
        uint32_t p=pixels[i]; unsigned a=p>>24,r=(p>>16)&255,g=(p>>8)&255,b=p&255;
        zeros+=a==0;nonopaque+=a!=255;
        invalid+=(r>a+1 || g>a+1 || b>a+1 || (a==0 && (r||g||b)));
    }
    free(pixels);
    int alpha=nonopaque>0;
    if(s->frames==0 || s->last_alpha!=alpha)
        fprintf(stderr,"client buffer surface=%u frame=%d size=%dx%d zero=%d nonopaque=%d invalid=%d\n",wl_resource_get_id(surface->resource),s->frames,width,height,zeros,nonopaque,invalid);
    s->frames++;s->last_alpha=alpha;
}
static void surface_created(struct wl_listener *listener,void *data) {
    struct observed_surface *s=calloc(1,sizeof(*s));if(!s)return;
    s->surface=data;s->commit.notify=observed_commit;s->destroy.notify=observed_destroy;
    wl_signal_add(&s->surface->commit_signal,&s->commit);wl_signal_add(&s->surface->destroy_signal,&s->destroy);
}
static int dispatch_control(int fd,uint32_t mask,void *data) {
    char message[80]={0};if(recv(fd,message,sizeof(message)-1,0)<=0)return 0;
    fprintf(stderr,"fixture command %s\n",message);
    struct bound_manager *m;struct surface_extension *e;
    if(!strcmp(message,"blur off") || !strcmp(message,"blur on")) {
        capabilities=!strcmp(message,"blur on");
        wl_list_for_each(m,&managers,link)ext_background_effect_manager_v1_send_capabilities(m->resource,capabilities);
    } else if(!strcmp(message,"remove") && background_global) {
        wl_global_destroy(background_global);background_global=NULL;
    } else if(!strcmp(message,"add") && !background_global) {
        capabilities=1;background_global=wl_global_create(compositor->wl_display,&ext_background_effect_manager_v1_interface,1,NULL,bind_background);
    } else if(sscanf(message,"scale %u",&preferred_scale)==1 && preferred_scale>0) {
        wl_list_for_each(e,&extensions,link)if(e->fractional)wp_fractional_scale_v1_send_preferred_scale(e->resource,preferred_scale);
    }
    return 0;
}
__attribute__((visibility("default"))) int wet_module_init(struct weston_compositor *c,int *argc,char *argv[]) {
    const char *path=getenv("FYNE_TEST_BACKGROUND_SOCKET");if(!path)return -1;
    compositor=c;wl_list_init(&managers);wl_list_init(&extensions);
    background_global=wl_global_create(c->wl_display,&ext_background_effect_manager_v1_interface,1,NULL,bind_background);
    fractional_global=wl_global_create(c->wl_display,&wp_fractional_scale_manager_v1_interface,1,NULL,bind_fractional);
    created.notify=surface_created;wl_signal_add(&c->create_surface_signal,&created);
    int fd=socket(AF_UNIX,SOCK_DGRAM|SOCK_CLOEXEC,0);if(fd<0)return -1;
    struct sockaddr_un address={.sun_family=AF_UNIX};if(strlen(path)>=sizeof(address.sun_path))return -1;
    strcpy(address.sun_path,path);if(bind(fd,(struct sockaddr*)&address,sizeof(address)))return -1;
    if(!wl_event_loop_add_fd(wl_display_get_event_loop(c->wl_display),fd,WL_EVENT_READABLE,dispatch_control,NULL))return -1;
    fprintf(stderr,"owned background fixture ready (protocol validation, no blur renderer)\n");return 0;
}
