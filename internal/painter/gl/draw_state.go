package gl

// drawBindings tracks state within one frame, while the painter owns the
// context. Material passes and new frames invalidate these bindings.
type drawBindings struct {
	program       Program
	buffer        Buffer
	blend         [2]uint32
	haveProgram   bool
	haveBuffer    bool
	haveBlend     bool
	vertexFormats map[Attribute]vertexFormat
}

type vertexFormat struct {
	buffer               Buffer
	size, stride, offset int
}

func (s *drawBindings) reset() {
	s.haveProgram, s.haveBuffer, s.haveBlend = false, false, false
	clear(s.vertexFormats)
}

func (p *painter) bindProgram(program Program) {
	s := &p.bindings
	if s.haveProgram && sameProgram(s.program, program) {
		return
	}
	p.ctx.UseProgram(program)
	s.program, s.haveProgram = program, true
}

func (p *painter) bindArrayBuffer(buffer Buffer) {
	s := &p.bindings
	if s.haveBuffer && sameBuffer(s.buffer, buffer) {
		return
	}
	p.ctx.BindBuffer(arrayBuffer, buffer)
	s.buffer, s.haveBuffer = buffer, true
}

func (p *painter) blendFunc(src, dst uint32) {
	s := &p.bindings
	blend := [2]uint32{src, dst}
	if s.haveBlend && s.blend == blend {
		return
	}
	p.ctx.BlendFunc(src, dst)
	s.blend, s.haveBlend = blend, true
}

func (p *painter) setVertexFormat(a Attribute, size, stride, offset int) {
	s := &p.bindings
	if !s.haveBuffer {
		// An untracked binding cannot be used to identify the input format.
		p.ctx.EnableVertexAttribArray(a)
		p.ctx.VertexAttribPointerWithOffset(a, size, float, false, stride, offset)
		return
	}
	if previous, ok := s.vertexFormats[a]; ok {
		if sameBuffer(previous.buffer, s.buffer) &&
			previous.size == size && previous.stride == stride && previous.offset == offset {
			return
		}
	} else {
		p.ctx.EnableVertexAttribArray(a)
		if s.vertexFormats == nil {
			s.vertexFormats = make(map[Attribute]vertexFormat)
		}
	}
	p.ctx.VertexAttribPointerWithOffset(a, size, float, false, stride, offset)
	s.vertexFormats[a] = vertexFormat{buffer: s.buffer, size: size, stride: stride, offset: offset}
}
