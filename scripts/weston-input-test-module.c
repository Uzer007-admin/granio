// Test-only input module for the owned Weston instance. notify_* traverses the
// normal compositor hit testing, implicit grab and serial machinery. The
// independent wl_data_device client remains responsible for start_drag/data.
#define _GNU_SOURCE
#include <libweston/libweston.h>
#include <wayland-server-core.h>
#include <linux/input-event-codes.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <time.h>
#include <string.h>
// Exported libweston-10 input ABI (the backend declarations are private).
void notify_motion_absolute(struct weston_seat*,const struct timespec*,double,double);
void notify_button(struct weston_seat*,const struct timespec*,int32_t,enum wl_pointer_button_state);
void notify_pointer_frame(struct weston_seat*);
static struct weston_seat*seat;
static int dispatch(int fd,uint32_t mask,void*data){char message[128]={0};ssize_t n=recv(fd,message,sizeof(message)-1,0);if(n<=0)return 0;struct timespec time;clock_gettime(CLOCK_MONOTONIC,&time);double x,y;int state;if(sscanf(message,"motion %lf %lf",&x,&y)==2)notify_motion_absolute(seat,&time,x,y);else if(sscanf(message,"button %d",&state)==1)notify_button(seat,&time,BTN_LEFT,state?WL_POINTER_BUTTON_STATE_PRESSED:WL_POINTER_BUTTON_STATE_RELEASED);notify_pointer_frame(seat);return 0;}
__attribute__((visibility("default"))) int wet_module_init(struct weston_compositor*compositor,int*argc,char*argv[]){const char*path=getenv("FYNE_WESTON_INPUT_SOCKET");if(!path)return -1;if(wl_list_empty(&compositor->seat_list))return -1;seat=wl_container_of(compositor->seat_list.next,seat,link);int fd=socket(AF_UNIX,SOCK_DGRAM|SOCK_CLOEXEC,0);if(fd<0)return -1;struct sockaddr_un address={.sun_family=AF_UNIX};if(strlen(path)>=sizeof(address.sun_path))return -1;strcpy(address.sun_path,path);if(bind(fd,(struct sockaddr*)&address,sizeof(address)))return -1;struct wl_event_loop*loop=wl_display_get_event_loop(compositor->wl_display);if(!wl_event_loop_add_fd(loop,fd,WL_EVENT_READABLE,dispatch,NULL))return -1;fprintf(stderr,"owned compositor input fixture ready\n");return 0;}
