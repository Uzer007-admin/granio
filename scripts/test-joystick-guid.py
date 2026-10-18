"""Compile the actual production GUID formatting blocks with boundary controls."""

import argparse
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent.parent


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cc", default="cc")
    parser.add_argument("--sanitize", action="store_true")
    parser.add_argument("--temp-root", type=Path)
    args = parser.parse_args()
    tests = []
    legacy = []
    for version in ("v3.3", "v3.4", "v3.5"):
        for platform, filename in (("linux", "linux_joystick.c"), ("windows", "win32_joystick.c"), ("macos", "cocoa_joystick.m")):
            path = ROOT / "deps/glfw" / version / "glfw/glfw/src" / filename
            source = path.read_text(encoding="utf-8")
            match = re.search(r'// Generate a joystick GUID.*?\n(    if .*?\n    }\n    else\n    {.*?\n    })', source, re.S)
            if not match:
                raise SystemExit(f"GUID block not found: {path}")
            declarations = list(re.finditer(r'char name\[256\][^;]*;',source[:match.start()]))
            if not declarations or declarations[-1].group() != 'char name[256] = "";':
                raise SystemExit(f"Actual GUID producer name bytes are not initialized: {path}")
            block = match.group(1)
            if 'sprintf(guid,' in block or block.count('snprintf(guid, sizeof(guid),') != 2:
                raise SystemExit(f"Unbounded GUID block: {path}")
            # sizeof(guid) must use the production array, not a decayed parameter.
            setup = '''struct { unsigned short bustype, vendor, product, version; } id;
                struct { struct { unsigned long Data1; unsigned char Data4[8]; } guidProduct; } instance, *di=&instance;
                int vendor=0, product=0, version=0;
                char name[256]=""; struct { unsigned char pre[16]; char guid[33]; unsigned char post[16]; } guarded;
                memset(&guarded,0xa5,sizeof(guarded)); char expected[33];
                unsigned char raw[16]={0}; id.bustype=3; id.vendor=id.product=id.version=0;
                memset(&instance,0,sizeof(instance));
            '''
            block = block.replace('guid, sizeof(guid)', 'guarded.guid, sizeof(guarded.guid)')
            # Golden bytes are constructed independently of printf.
            body = setup + '\nfor(int test=0;test<11;test++){\nmemset(name,0,sizeof(name));\n'
            body += 'if(test==0)memcpy(name,"ASCII",5); if(test==1)memcpy(name,"ABC\\xc3\\xa9",5); if(test==2)memset(name,0xff,11); if(test==3)name[0]=(char)0x80; if(test>=4)memset(name,0x7f,11);\n'
            body += 'if(test>=9){memset(name,0,sizeof(name));if(test==10)name[0]=65;}\n'
            if platform == "linux":
                body += 'id.vendor=test>=4?0xffff:0; id.product=test>=4?0xffff:0; id.version=test>=4?0xffff:0; if(test==5)id.vendor=0; if(test==6)id.product=0; if(test==7)id.version=0; if(test==8)memset(name,0,sizeof(name));if(test>=9)id.vendor=id.product=id.version=0;\n'
                body += 'memset(raw,0,16); raw[0]=3; if(id.vendor&&id.product&&id.version){raw[4]=raw[5]=raw[8]=raw[9]=raw[12]=raw[13]=255;}else memcpy(raw+4,name,11);\n'
            else:
                body += 'vendor=product=version=test>=4?0xffff:0; instance.guidProduct.Data1=0xffffffff; if(test>=4)memcpy(instance.guidProduct.Data4+2,"PIDVID",6); if(test==5){vendor=0;memset(instance.guidProduct.Data4,0,8);} if(test==6){product=0;memset(instance.guidProduct.Data4,0,8);} if(test==8)memset(name,0,sizeof(name));if(test>=9){vendor=product=version=0;memset(instance.guidProduct.Data4,0,8);}\n'
                predicate = 'memcmp(di->guidProduct.Data4+2,"PIDVID",6)==0' if platform == 'windows' else 'vendor&&product'
                body += 'memset(raw,0,16); if('+predicate+'){raw[0]=3;raw[4]=raw[5]=raw[8]=raw[9]=255;'
                if platform == 'macos': body += 'raw[12]=raw[13]=255;'
                body += '}else{raw[0]=5;memcpy(raw+4,name,11);}\n'
            body = body.replace('; if', ';\nif').replace(';if', ';\nif')
            body += block + '\n'
            body += '(void)id; (void)di; (void)vendor; (void)product; (void)version;\n'
            body += '''static const char hex[]="0123456789abcdef";
                for(int i=0;i<16;i++){expected[2*i]=hex[raw[i]>>4];expected[2*i+1]=hex[raw[i]&15];} expected[32]=0;
                if(memcmp(guarded.guid,expected,33)){fprintf(stderr,"case %d: %s != %s\\n",test,guarded.guid,expected);return 1;}
                for(int i=0;i<16;i++)if(guarded.pre[i]!=0xa5||guarded.post[i]!=0xa5)return 2;
            } return 0;
            '''
            name = version.replace('.', '_') + '_' + platform
            tests.append((name, body))
            conversion = re.search(r'void (_glfw\w*UpdateGamepadGUID\w*)\(char\* guid\)\n({.*?\n})', source, re.S)
            if not conversion:
                raise SystemExit(f"Legacy GUID converter not found: {path}")
            if 'sizeof(guid)' in conversion.group(2):
                raise SystemExit(f"sizeof(pointer) in legacy converter: {path}")
            old_name = name + '_legacy'
            if platform == 'windows':
                before='12345678000000000000504944564944'; after='03000000123400005678000000000000'
            elif platform == 'macos':
                before='12340000000000005678000000000000'; after='03000000123400005678000000000000'
            else:
                before=after='03000000123400005678000000000000'
            legacy.append('static void '+old_name+'(char* guid)\n'+conversion.group(2).replace('{', '{ (void)guid;',1)+'\n')
            legacy.append('static int '+old_name+'_test(void){struct {unsigned char pre[16];char guid[33];unsigned char post[16];} g;memset(&g,0xa5,sizeof(g));memcpy(g.guid,"'+before+'",33);'+old_name+'(g.guid);if(memcmp(g.guid,"'+after+'",33))return 1;for(int i=0;i<16;i++)if(g.pre[i]!=0xa5||g.post[i]!=0xa5)return 2;return 0;}\n')
    code = '#include <stdio.h>\n#include <string.h>\n#include <stdint.h>\n'
    code += ''.join('static int '+name+'(void){\n'+body+'}\n' for name,body in tests)
    code += ''.join(legacy)
    code += 'int main(void){\n'+''.join('if('+name+'()||'+name+'_legacy_test())return 1;\n' for name,_ in tests)+'puts("9 production blocks x 11 controls + 9 legacy converters passed");return 0;}\n'
    with tempfile.TemporaryDirectory(prefix="guid-regression-", dir=args.temp_root) as temporary:
        source=Path(temporary)/'guid.c'; source.write_text(code,encoding='utf-8')
        for char in ('signed','unsigned'):
            output=Path(temporary)/('guid-'+char+'.exe')
            command=[args.cc,'-std=c99','-Wall','-Wextra','-Werror','-f'+char+'-char',str(source),'-o',str(output)]
            if args.sanitize: command[1:1]=['-fsanitize=address,undefined','-fno-omit-frame-pointer']
            subprocess.run(command,check=True)
            subprocess.run([str(output)],check=True)
            print(char+' char: passed',flush=True)


if __name__ == '__main__':
    main()
