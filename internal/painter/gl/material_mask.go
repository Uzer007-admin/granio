//go:build !android && !ios && !mobile && !wasm && !test_web_driver

package gl

func (n *nativeMaterialPipeline) ensureMask(p *painter) bool {
	if n.mask.ref != noProgram {
		return true
	}
	prefix := "#version 110\n"
	if n.es {
		prefix = "#version 100\nprecision highp float;\n"
	}
	program, err := p.createProgram(shaderVertTexturedPassthrough2D, []byte(prefix+materialCoverageShader))
	if err != nil {
		return false
	}
	n.mask = programState{ref: program, buff: p.createBuffer(coordinatesSizeRectangleWithTexture), uniforms: make(map[string]*uniformState), attributes: make(map[string]Attribute)}
	return true
}

// Apply the same full-resolution coverage when erasing and when adding the
// filtered snapshot. Its premultiplied alpha is preserved at mixed corners.
const materialCoverageShader = `
uniform vec2 extent;
uniform vec4 radii; // TL, TR, BR, BL
uniform sampler2D tex;
uniform float pass;
varying vec2 fragTexCoord;
void main() {
    vec2 p=vec2(fragTexCoord.x,1.0-fragTexCoord.y);
    float r=p.y<0.5?(p.x<0.5?radii.x:radii.y):(p.x<0.5?radii.w:radii.z);
    r=min(r,min(extent.x,extent.y)*0.5);
    vec2 q=abs((p-0.5)*extent)-extent*0.5+r;
    float d=length(max(q,vec2(0.0)))+min(max(q.x,q.y),0.0)-r;
    float a=clamp(0.5-d,0.0,1.0);
    gl_FragColor=pass>0.5?texture2D(tex,fragTexCoord)*a:vec4(0.0,0.0,0.0,a);
}
`
