#ifndef GRANIO_SOFTWARE_LINUX_H
#define GRANIO_SOFTWARE_LINUX_H

#include <stdint.h>

typedef struct GranioSoftwareSurface GranioSoftwareSurface;

enum {
    GRANIO_SOFTWARE_PRESENTED,
    GRANIO_SOFTWARE_DEFERRED,
    GRANIO_SOFTWARE_UNSUPPORTED,
    GRANIO_SOFTWARE_FAILED
};

GranioSoftwareSurface* granioCreateSoftwareSurface(void);
void granioDestroySoftwareSurface(GranioSoftwareSurface*);
int granioPresentSoftwareX11(GranioSoftwareSurface*, void* display, uintptr_t window,
                             const unsigned char*, int width, int height, int stride);
int granioPresentSoftwareWayland(GranioSoftwareSurface*, void* display, void* surface,
                                 const unsigned char*, int width, int height, int stride,
                                 int surfaceWidth, int surfaceHeight);

#endif
