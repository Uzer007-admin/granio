// An independent wl_data_device source. A real compositor button serial starts
// the drag; no private Fyne callback or synthetic Wayland serial is used.
#define _GNU_SOURCE
#include <wayland-client.h>
#include "xdg-shell-client-protocol.h"
#include <linux/input-event-codes.h>
#include <sys/mman.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <errno.h>
static struct wl_display*display;static struct wl_compositor*compositor;static struct wl_shm*shm;static struct wl_seat*seat;static struct xdg_wm_base*shell;static struct wl_data_device_manager*manager;static struct wl_data_device*device;static struct wl_data_source*source;static struct wl_surface*surface;static struct wl_buffer*buffer;
static char*mime;static char*payload;static size_t length;static int done;
static struct wl_data_offer*selfOffer;
static void sent(void*d,struct wl_data_source*s,const char*type,int fd){if(strcmp(type,mime)){fprintf(stderr,"unadvertised MIME requested: %s\n",type);close(fd);done=2;return;}size_t pos=0;while(pos<length){ssize_t n=write(fd,payload+pos,length-pos);if(n<0&&errno==EINTR)continue;if(n<=0)break;pos+=n;}close(fd);printf("sent:%s:%zu\n",type,pos);fflush(stdout);}
static void target(void*d,struct wl_data_source*s,const char*type){printf("target:%s\n",type?type:"none");fflush(stdout);}
static void cancelled(void*d,struct wl_data_source*s){done=2;puts("cancelled");}
static void dropped(void*d,struct wl_data_source*s){puts("drop-performed");}
static void finished(void*d,struct wl_data_source*s){done=1;puts("finished");}
static void action(void*d,struct wl_data_source*s,uint32_t action){printf("action:%u\n",action);}
static const struct wl_data_source_listener sourceListener={target,sent,cancelled,dropped,finished,action};
static void offerType(void*d,struct wl_data_offer*o,const char*type){}
static void offerActions(void*d,struct wl_data_offer*o,uint32_t actions){}
static const struct wl_data_offer_listener offerListener={offerType,offerActions,offerActions};
static void offer(void*d,struct wl_data_device*v,struct wl_data_offer*o){wl_data_offer_add_listener(o,&offerListener,NULL);}
static void deviceEnter(void*d,struct wl_data_device*v,uint32_t serial,struct wl_surface*s,wl_fixed_t x,wl_fixed_t y,struct wl_data_offer*o){printf("source-device-enter:%u\n",serial);fflush(stdout);selfOffer=o;if(o){wl_data_offer_accept(o,serial,NULL);wl_data_offer_set_actions(o,WL_DATA_DEVICE_MANAGER_DND_ACTION_COPY,WL_DATA_DEVICE_MANAGER_DND_ACTION_COPY);}}
static void deviceLeave(void*d,struct wl_data_device*v){if(selfOffer){wl_data_offer_destroy(selfOffer);selfOffer=NULL;}}
static void deviceMotion(void*d,struct wl_data_device*v,uint32_t time,wl_fixed_t x,wl_fixed_t y){}
static void deviceDrop(void*d,struct wl_data_device*v){}
static void deviceSelection(void*d,struct wl_data_device*v,struct wl_data_offer*o){if(o)wl_data_offer_destroy(o);}
static const struct wl_data_device_listener deviceListener={offer,deviceEnter,deviceLeave,deviceMotion,deviceDrop,deviceSelection};
static void enter(void*d,struct wl_pointer*p,uint32_t serial,struct wl_surface*s,wl_fixed_t x,wl_fixed_t y){printf("pointer-enter:%u\n",serial);fflush(stdout);}
static void leave(void*d,struct wl_pointer*p,uint32_t serial,struct wl_surface*s){}
static void motion(void*d,struct wl_pointer*p,uint32_t time,wl_fixed_t x,wl_fixed_t y){}
static void button(void*d,struct wl_pointer*p,uint32_t serial,uint32_t time,uint32_t key,uint32_t state){if(key==BTN_LEFT&&state==WL_POINTER_BUTTON_STATE_PRESSED){source=wl_data_device_manager_create_data_source(manager);wl_data_source_add_listener(source,&sourceListener,NULL);wl_data_source_offer(source,mime);wl_data_source_set_actions(source,WL_DATA_DEVICE_MANAGER_DND_ACTION_COPY);wl_data_device_start_drag(device,source,surface,NULL,serial);wl_display_flush(display);puts("started");fflush(stdout);}}
static void axis(void*d,struct wl_pointer*p,uint32_t time,uint32_t axis,wl_fixed_t value){}
static const struct wl_pointer_listener pointerListener={enter,leave,motion,button,axis};
static void capabilities(void*d,struct wl_seat*s,uint32_t caps){printf("seat:%u\n",caps);fflush(stdout);if(caps&WL_SEAT_CAPABILITY_POINTER){struct wl_pointer*p=wl_seat_get_pointer(s);wl_pointer_add_listener(p,&pointerListener,NULL);}}
static const struct wl_seat_listener seatListener={capabilities};
static void ping(void*d,struct xdg_wm_base*s,uint32_t serial){xdg_wm_base_pong(s,serial);}
static const struct xdg_wm_base_listener shellListener={ping};
static void global(void*d,struct wl_registry*r,uint32_t name,const char*type,uint32_t version){if(!strcmp(type,"wl_compositor"))compositor=wl_registry_bind(r,name,&wl_compositor_interface,1);else if(!strcmp(type,"wl_shm"))shm=wl_registry_bind(r,name,&wl_shm_interface,1);else if(!strcmp(type,"wl_seat")){seat=wl_registry_bind(r,name,&wl_seat_interface,1);wl_seat_add_listener(seat,&seatListener,NULL);}else if(!strcmp(type,"wl_data_device_manager"))manager=wl_registry_bind(r,name,&wl_data_device_manager_interface,version<3?version:3);else if(!strcmp(type,"xdg_wm_base")){shell=wl_registry_bind(r,name,&xdg_wm_base_interface,1);xdg_wm_base_add_listener(shell,&shellListener,NULL);}}
static void removed(void*d,struct wl_registry*r,uint32_t name){}
static const struct wl_registry_listener registryListener={global,removed};
static void configure(void*d,struct xdg_surface*s,uint32_t serial){xdg_surface_ack_configure(s,serial);wl_surface_attach(surface,buffer,0,0);wl_surface_damage(surface,0,0,120,120);wl_surface_commit(surface);puts("ready");fflush(stdout);}
static const struct xdg_surface_listener surfaceListener={configure};
static void topConfigure(void*d,struct xdg_toplevel*t,int32_t w,int32_t h,struct wl_array*states){}
static void topClose(void*d,struct xdg_toplevel*t){done=2;}
static const struct xdg_toplevel_listener topListener={topConfigure,topClose};
int main(int argc,char**argv){if(argc!=3)return 2;signal(SIGPIPE,SIG_IGN);alarm(20);mime=argv[1];FILE*f=fopen(argv[2],"rb");if(!f)return 3;fseek(f,0,SEEK_END);length=ftell(f);rewind(f);payload=malloc(length?length:1);if(fread(payload,1,length,f)!=length)return 4;fclose(f);display=wl_display_connect(NULL);if(!display)return 5;struct wl_registry*r=wl_display_get_registry(display);wl_registry_add_listener(r,&registryListener,NULL);wl_display_roundtrip(display);wl_display_roundtrip(display);if(!compositor||!shm||!seat||!shell||!manager||wl_data_device_manager_get_version(manager)<3)return 6;device=wl_data_device_manager_get_data_device(manager,seat);wl_data_device_add_listener(device,&deviceListener,NULL);surface=wl_compositor_create_surface(compositor);
int fd=memfd_create("fyne-independent-drag",0);if(fd<0||ftruncate(fd,120*120*4))return 7;uint32_t*pixels=mmap(NULL,120*120*4,PROT_READ|PROT_WRITE,MAP_SHARED,fd,0);for(int i=0;i<120*120;i++)pixels[i]=0xffeb2142;struct wl_shm_pool*pool=wl_shm_create_pool(shm,fd,120*120*4);buffer=wl_shm_pool_create_buffer(pool,0,120,120,120*4,WL_SHM_FORMAT_XRGB8888);wl_shm_pool_destroy(pool);close(fd);struct xdg_surface*xs=xdg_wm_base_get_xdg_surface(shell,surface);xdg_surface_add_listener(xs,&surfaceListener,NULL);struct xdg_toplevel*top=xdg_surface_get_toplevel(xs);xdg_toplevel_add_listener(top,&topListener,NULL);xdg_toplevel_set_title(top,"Independent Wayland drag source");xdg_toplevel_set_app_id(top,"org.fyne.independent-drag");wl_surface_commit(surface);while(!done&&wl_display_dispatch(display)>=0){}fflush(stdout);wl_display_disconnect(display);free(payload);munmap(pixels,120*120*4);return done==1?0:8;}
