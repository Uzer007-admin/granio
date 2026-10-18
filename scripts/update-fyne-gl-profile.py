"""Regenerate the GL 2.1 symbol restriction from Fyne's native GL adapter."""

import json
from pathlib import Path
import re
import xml.etree.ElementTree as ET


ROOT = Path(__file__).resolve().parent.parent


def main():
    registry = ET.parse(ROOT / "deps/glow/xml/spec/gl.xml").getroot()
    functions = {}
    for command in registry.findall("./commands/command"):
        name = command.findtext("./proto/name") or command.get("name")
        if name and name.startswith("gl"):
            functions[name[2:]] = name
    for overload in ET.parse(ROOT / "deps/glow/xml/overload/gl.xml").getroot():
        functions[overload.get("overloadName")] = overload.get("name")
    enum_names = {enum.get("name") for enum in registry.findall("./enums/enum")}
    source = (ROOT / "internal/painter/gl/gl_core.go").read_text(encoding="utf-8")
    selectors = set(re.findall(r"\b(?:gl|procs|c\.procs)\.([A-Za-z_][A-Za-z_0-9]*)", source))
    # The length-delimited helper still consumes the native glShaderSource entry.
    selectors = {"ShaderSource" if symbol == "ShaderSourceString" else symbol for symbol in selectors}
    helpers = {"Functions", "Load", "LoadWithProcAddrFunc", "Init", "InitWithProcAddrFunc", "IsFunctionAvailable", "Ptr", "SlicePtr", "SliceBytes", "Str", "Strs", "StrsWithLengths", "GoStr"}
    enums, commands = set(), {"glGetString", "glDeleteShader"}  # Renderer/version and shader diagnostics.
    for symbol in sorted(selectors - helpers):
        if "GL_" + symbol in enum_names:
            enums.add("GL_" + symbol)
        elif symbol in functions:
            commands.add(functions[symbol])
        else:
            raise ValueError(f"Unrecognized GL selector: {symbol}")
    restriction = {"enums": sorted(enums | {"GL_VERSION", "GL_RENDERER", "GL_TRUE"}), "functions": sorted(commands)}
    target = ROOT / "deps/gl/profiles/fyne-gl21.json"
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(json.dumps(restriction, indent=2) + "\n", encoding="utf-8")
    print(f"{len(restriction['functions'])} functions, {len(restriction['enums'])} enums: {target}")


if __name__ == "__main__":
    main()
