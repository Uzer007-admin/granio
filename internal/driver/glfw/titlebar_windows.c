//go:build windows && !wasm && !test_web_driver

#include <windows.h>
#include <windowsx.h>
#include <stdlib.h>

typedef struct {
    int height, leading, trailing;
    int border_x, border_y;
    int min_width, min_height, max_width, max_height;
    WNDPROC previous;
} GranioTitleBar;

static const wchar_t title_bar_property[] = L"Granio.ClientTitleBar";

static GranioTitleBar title_bar_geometry(HWND hwnd, int height, int leading, int trailing) {
    typedef UINT (WINAPI *GetDpiFn)(HWND);
    typedef int (WINAPI *GetMetricsFn)(int, UINT);
    HMODULE user32 = GetModuleHandleW(L"user32.dll");
    GetDpiFn get_dpi = (GetDpiFn)GetProcAddress(user32, "GetDpiForWindow");
    GetMetricsFn get_metrics = (GetMetricsFn)GetProcAddress(user32, "GetSystemMetricsForDpi");
    UINT dpi = get_dpi ? get_dpi(hwnd) : 0;
    GranioTitleBar result = {.height = height, .leading = leading, .trailing = trailing};
    if (get_metrics && dpi) {
        int padding = get_metrics(SM_CXPADDEDBORDER, dpi);
        result.border_x = get_metrics(SM_CXSIZEFRAME, dpi) + padding;
        result.border_y = get_metrics(SM_CYSIZEFRAME, dpi) + padding;
    } else {
        int padding = GetSystemMetrics(SM_CXPADDEDBORDER);
        result.border_x = GetSystemMetrics(SM_CXSIZEFRAME) + padding;
        result.border_y = GetSystemMetrics(SM_CYSIZEFRAME) + padding;
    }
    return result;
}

static LRESULT title_bar_hit(HWND hwnd, LPARAM lparam, LONG_PTR style, const GranioTitleBar *bar) {
    RECT client;
    POINT cursor = { GET_X_LPARAM(lparam), GET_Y_LPARAM(lparam) };
    if (!GetClientRect(hwnd, &client) || !ScreenToClient(hwnd, &cursor) || !PtInRect(&client, cursor))
        return HTNOWHERE;
    if ((style & WS_THICKFRAME) && !IsZoomed(hwnd)) {
        int x_border = bar->border_x, y_border = bar->border_y;
        BOOL left = cursor.x < x_border, right = cursor.x >= client.right - x_border;
        BOOL top = cursor.y < y_border, bottom = cursor.y >= client.bottom - y_border;
        if (top) return left ? HTTOPLEFT : right ? HTTOPRIGHT : HTTOP;
        if (bottom) return left ? HTBOTTOMLEFT : right ? HTBOTTOMRIGHT : HTBOTTOM;
        if (left) return HTLEFT;
        if (right) return HTRIGHT;
    }
    if (cursor.y < bar->height && cursor.x >= bar->leading && cursor.x < client.right - bar->trailing)
        return HTCAPTION;
    return HTCLIENT;
}

static LRESULT CALLBACK title_bar_proc(HWND hwnd, UINT msg, WPARAM wparam, LPARAM lparam) {
    GranioTitleBar *bar = (GranioTitleBar *)GetPropW(hwnd, title_bar_property);
    if (!bar)
        return DefWindowProcW(hwnd, msg, wparam, lparam);
    WNDPROC previous = bar->previous;
    if (msg == WM_NCDESTROY) {
        RemovePropW(hwnd, title_bar_property);
        free(bar);
    } else if (msg == WM_NCCALCSIZE || msg == WM_NCHITTEST || msg == WM_GETMINMAXINFO) {
        LONG_PTR style = GetWindowLongPtrW(hwnd, GWL_STYLE);
        if ((style & WS_CAPTION) == WS_CAPTION) {
            if (msg == WM_NCHITTEST)
                return title_bar_hit(hwnd, lparam, style, bar);
            if (msg == WM_GETMINMAXINFO) {
                LRESULT result = CallWindowProcW(previous, hwnd, msg, wparam, lparam);
                MINMAXINFO *limits = (MINMAXINFO *)lparam;
                if (bar->min_width > 0) limits->ptMinTrackSize.x = bar->min_width;
                if (bar->min_height > 0) limits->ptMinTrackSize.y = bar->min_height;
                if (bar->max_width > 0) limits->ptMaxTrackSize.x = bar->max_width;
                if (bar->max_height > 0) limits->ptMaxTrackSize.y = bar->max_height;
                return result;
            }
            if (IsZoomed(hwnd) && lparam) {
                RECT *rect = wparam ? &((NCCALCSIZE_PARAMS *)lparam)->rgrc[0] : (RECT *)lparam;
                MONITORINFO monitor = { .cbSize = sizeof(monitor) };
                RECT usable;
                if (GetMonitorInfoW(MonitorFromRect(rect, MONITOR_DEFAULTTONEAREST), &monitor) &&
                    IntersectRect(&usable, rect, &monitor.rcWork))
                    *rect = usable;
            }
            return 0;
        }
    }
    return CallWindowProcW(previous, hwnd, msg, wparam, lparam);
}

int granio_client_title_bar(HWND hwnd, int height, int leading, int trailing) {
    GranioTitleBar *bar = (GranioTitleBar *)GetPropW(hwnd, title_bar_property);
    GranioTitleBar updated = title_bar_geometry(hwnd, height, leading, trailing);
    if (bar) {
        bar->height = updated.height;
        bar->leading = updated.leading;
        bar->trailing = updated.trailing;
        bar->border_x = updated.border_x;
        bar->border_y = updated.border_y;
        return 1;
    }
    bar = malloc(sizeof(*bar));
    if (!bar) return 0;
    *bar = updated;
    if (SetPropW(hwnd, title_bar_property, bar)) {
        bar->previous = (WNDPROC)SetWindowLongPtrW(hwnd, GWLP_WNDPROC, (LONG_PTR)title_bar_proc);
        if (bar->previous) {
            WNDPROC previous = bar->previous;
            BOOL resized = SetWindowPos(hwnd, NULL, 0, 0, 0, 0,
                                        SWP_NOMOVE | SWP_NOSIZE | SWP_NOZORDER | SWP_NOACTIVATE | SWP_FRAMECHANGED);
            if (GetPropW(hwnd, title_bar_property) != bar)
                return 0;
            if (resized)
                return 1;
            SetWindowLongPtrW(hwnd, GWLP_WNDPROC, (LONG_PTR)previous);
        }
        RemovePropW(hwnd, title_bar_property);
    }
    free(bar);
    return 0;
}

int granio_title_bar_bounds(HWND hwnd, int first, int second, int position) {
    UINT flags = SWP_NOACTIVATE | SWP_NOZORDER | SWP_NOOWNERZORDER;
    if (position)
        return SetWindowPos(hwnd, NULL, first, second, 0, 0, flags | SWP_NOSIZE);
    return SetWindowPos(hwnd, NULL, 0, 0, first, second, flags | SWP_NOMOVE);
}

int granio_title_bar_limits(HWND hwnd, int min_width, int min_height, int max_width, int max_height) {
    GranioTitleBar *bar = (GranioTitleBar *)GetPropW(hwnd, title_bar_property);
    if (!bar) return 0;
    bar->min_width = min_width;
    bar->min_height = min_height;
    bar->max_width = max_width;
    bar->max_height = max_height;
    return 1;
}
