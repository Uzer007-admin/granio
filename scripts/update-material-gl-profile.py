"""Build the optional GL 2.1 material profile without raising Fyne's baseline."""

import json
from pathlib import Path
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parent.parent


def main():
    source = json.loads((ROOT / "deps/gl/profiles/fyne-gl21.json").read_text())
    functions, enums = set(source["functions"]), set(source["enums"])
    functions.update(("glBlendFuncSeparate", "glBlendFuncSeparateEXT", "glColorMask", "glGetBooleanv"))
    for name in ("GenFramebuffers", "DeleteFramebuffers", "BindFramebuffer", "FramebufferTexture2D", "CheckFramebufferStatus", "GetFramebufferAttachmentParameteriv", "GenRenderbuffers", "DeleteRenderbuffers", "BindRenderbuffer", "RenderbufferStorage", "FramebufferRenderbuffer", "GetRenderbufferParameteriv"):
        functions.update(("gl" + name, "gl" + name + "EXT"))
    for name in ("FRAMEBUFFER", "FRAMEBUFFER_BINDING", "FRAMEBUFFER_COMPLETE", "COLOR_ATTACHMENT0", "DEPTH_ATTACHMENT", "STENCIL_ATTACHMENT", "RENDERBUFFER", "RENDERBUFFER_BINDING", "MAX_RENDERBUFFER_SIZE"):
        enums.update(("GL_" + name, "GL_" + name + "_EXT"))
    enums.update("GL_" + name for name in ("RGBA8", "DEPTH_COMPONENT16", "ALPHA_BITS", "RED_BITS", "GREEN_BITS", "BLUE_BITS", "SAMPLES", "SAMPLE_BUFFERS", "EXTENSIONS", "VENDOR"))
    registry = ET.parse(ROOT / "deps/glow/xml/spec/gl.xml").getroot()
    known_functions = {node.findtext("./proto/name") or node.get("name") for node in registry.findall("./commands/command")}
    known_enums = {node.get("name") for node in registry.findall("./enums/enum")}
    if functions - known_functions or enums - known_enums:
        raise ValueError((functions - known_functions, enums - known_enums))
    target = ROOT / "deps/gl/profiles/material-gl21.json"
    target.write_text(json.dumps({"functions": sorted(functions), "enums": sorted(enums)}, indent=2) + "\n", encoding="utf-8")
    print(f"Material profile: {len(functions)} functions, {len(enums)} enums")


if __name__ == "__main__":
    main()
