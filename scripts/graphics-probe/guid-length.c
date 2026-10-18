// Safe reproduction of go-gl/glfw#426: measure output in a large buffer.
#include <stdio.h>

int main(void) {
    char name[256] = "ABC\xc3\xa9";
    char output[128];
    const char *format = "%02x%02x0000%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x00";
    int length = snprintf(output, sizeof(output), format, 3, 0,
                          name[0], name[1], name[2], name[3], name[4],
                          name[5], name[6], name[7], name[8], name[9], name[10]);
    int masked = snprintf(output, sizeof(output), format, 3, 0,
                          name[0] & 0xff, name[1] & 0xff, name[2] & 0xff,
                          name[3] & 0xff, name[4] & 0xff, name[5] & 0xff,
                          name[6] & 0xff, name[7] & 0xff, name[8] & 0xff,
                          name[9] & 0xff, name[10] & 0xff);
    printf("char_signed=%d original_length=%d bytes_with_nul=%d guid_capacity=33 masked_length=%d\n",
           (char)0xff < 0, length, length + 1, masked);
    return length != 44 || masked != 32;
}
