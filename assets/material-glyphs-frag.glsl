#version 100
precision mediump float;

uniform sampler2D tex0;
varying vec2 vtc0;

uniform vec2 icon;
uniform vec4 color;
varying vec3 vpos;
varying float vheight;

const float hglyph = 72.0;
const float spread = 2.0;
const float edge = 0.5;

const float scale = 64.0/50.0; // also see Text.Draw

void main() {
  if (icon.x >= 0.0) {
    float acoef = 0.87;
    vec2 tc = vtc0+icon;
    tc.y -= hglyph/512.0/scale;
    float d = texture2D(tex0, tc).a;
    float h = vheight/1.2;
    float gnum = 0.25; // just a numerator
    if (11.5 < h && h <= 12.5) {
      gnum = 0.105;
      d += 0.2;
      acoef *= 0.87;
    } else if (12.5 < h && h <= 14.5) {
      gnum = 0.15;
      d += 0.08;
    } else if (14.5 < h && h <= 18.5) {
      gnum = 0.15;
      d += 0.04;
    } else if (18.5 < h && h <= 20.5) {
      gnum = 0.2;
    }
    float gamma = gnum/(spread*(h/hglyph)); 

    // d += 0.2; // bold
    // d -= 0.2; // thin

    gl_FragColor.rgb = color.rgb;
    gl_FragColor.a = smoothstep(edge-gamma, edge+gamma, d);
    gl_FragColor.a *= acoef; // secondary text

    // TODO
    // gl_FragColor = texture2D(tex0, vtc0+icon);
  } else {
    // gl_FragColor = color;
    gl_FragColor = vec4(1.0, 0.0, 0.0, 1.0);
  }
}
