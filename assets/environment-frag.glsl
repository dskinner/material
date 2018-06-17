#version 100
#define onesqrt2 0.70710678118
#define sqrt2 1.41421356237
#define pi 3.14159265359
#define twopi 6.28318530718

#define touchBegin 0.0
#define touchMove 1.0
#define touchEnd 2.0
precision mediump float;

// TODO pass this in some other way so sampler can be selected
// 0:fontsize, 1:pad, 2:edge
uniform vec4 glyphconf;

uniform sampler2D image;

uniform sampler2D texglyph;
uniform sampler2D texicon;
uniform vec2 glyph;
uniform vec4 shadowColor;

varying vec4 vtexcoord;
varying vec4 vvertex;
varying vec4 vtouch;

// interpolated distance, and size values
// x, y is unit value [0..1]
// z, w is material width, height
varying vec4 vdist;
varying vec4 vcolor;

vec4 sampleIcon() {
  vec4 clr = texture2D(texicon, vtexcoord.xy);
  clr.rgb += vcolor.rgb;
  clr.a *= 0.54; // https://www.google.com/design/spec/style/color.html#color-ui-color-application
  return clr;
}

vec4 sampleImage() {
  vec4 clr = texture2D(image, vtexcoord.xy);
  return clr;
}

// four values for a texture might look like this
// [
//  morton-encoded bounds,
//  texture size,
//  ?type? so as to invoke specific shader bits (glyph, icon, image)
//  alpha
// ]
//
// numbers are 32-bit
// must have four values to id rectangle for texture coordinates.
// could reduce this to 3 numbers if two were uint16 containers and one
// specified texture size. If texture size could be set only once then
// it could be two. Perhaps set texture size in a uniform? that would
// allow for 2 numbers to specify rectangle and 1 for type and 1 for alpha.

vec4 sampleGlyph() {
  float fontsize = glyphconf.x;
  float pad = glyphconf.y;
  float edge = glyphconf.z;

  float d = texture2D(texglyph, vtexcoord.xy).a;
  float h = vdist.w;
  float gamma = 0.22/(pad*(h/fontsize));

  // d += 0.2; // bold
  // d -= 0.2; // thin

  vec4 clr = vcolor;
  clr.a = smoothstep(edge-gamma, edge+gamma, d);
  clr.a *= 0.87; // secondary text
  return clr;
}

// TODO drop this
bool shouldcolor(vec2 pos, float sz) {
  // maps 0.0 .. 0.5 .. 1.0
	//   to 0.0 .. 0.5 .. 0.0
  pos = 0.5-abs(pos-0.5);

	// multiply by width/height
  pos *= vdist.zw;

  // 
  if (pos.x <= sz && pos.y <= sz) {
    float d = length(1.0-(pos/sz));
    if (d > 1.0) {
      return false;
    }
  }
  return true;
}

float shade(vec2 pos, float sz) {
  pos = 0.5-abs(pos-0.5);
  pos *= vdist.zw;

  if (pos.x <= sz && pos.y <= sz) {
    float d = length(1.0-(pos/sz));
    if (d > 1.0) { // TODO consider moving this as discard into top of main
      return 0.0;
    }
    return 1.0-d;
  } else if (pos.x <= sz && pos.y > sz) {
    return pos.x/sz;
  } else if (pos.x > sz && pos.y <= sz) {
    return pos.y/sz;
  }
  return 1.0;
}

void main() {
  float roundness = vvertex.w;

	if (vtexcoord.x >= 0.0) {
    if (vtexcoord.z == 3.0) {
      gl_FragColor = sampleImage();
    } else if (vtexcoord.z == 1.0) {
      gl_FragColor = sampleIcon();
    } else {
      gl_FragColor = sampleGlyph();
    }
	} else if (vvertex.z <= 0.0) { // draw shadow
    if (roundness < 8.0) {
      roundness = 8.0;
    }

    // TODO ellipsis are being over-rounded resulting in a distortion
    // in the shadow. The distortion isn't very noticable unless drawing
    // only shadows, but this does affect outline of material making it
    // more difficult to see material edges. Clamping roundness results
    // in a strange distortion when animating roundness and the cause is
    // not currently clear.
    //
    // At the same time, the distortion looks better for different cases.
    roundness += -vvertex.z;
    // if (roundness > vdist.z/2.0) {
    // roundness = vdist.z/2.0;
    // }

		gl_FragColor = shadowColor;

    // maps roundness to 1 - [0..0.66] and helps shadows cast by rectangles and ellipses
    // look similar at the same z index.
    // TODO vvertex.w is roundness, double check usage here as was used before float roundness declared
    float e = 1.0 - (vvertex.w/vdist.z/0.75);

    gl_FragColor.a = smoothstep(0.0, e, shade(vdist.xy, roundness));
    vec2 n = abs(vdist.xy*2.0 - 1.0);
    n = n*n*n*n;
    n = 1.0-n;

    // reduce alpha/strength as z-index increases
    float f = 1.0 + (-vvertex.z*0.1);

    gl_FragColor.a *= n.x*n.y/f;
    // gl_FragColor.a *= 10.0;
  } else { // draw material
    if (shouldcolor(vdist.xy, roundness)) {
      gl_FragColor = vcolor;

      // anti-alias roundness
      if (gl_FragColor.a != 0.0) { // guard against exposing bg of text and icon content
        float dist = 1.0-shade(vdist.xy, roundness);
        // fractional based on largest size, approximates a consistent value across resolutions
        float dt = (5.0/max(vdist.z, vdist.w));
        gl_FragColor.a = 1.0-smoothstep(1.0-dt, 1.0, dist);
      }

      // respond to touch with radial
      const float dur = 200.0;
      const float clr = 0.035;

      float since = vtouch.w;
      vec4 react = vec4(0.0);

      // float d = length(vtouch.xy-vdist.xy);
      // float d = length(vtouch.xx-vdist.xx); // horizontal sweep

      const float magic = 100.0;
      float d = length((vtouch.xy*vdist.zw)-(vdist.xy*vdist.zw));
      d /= magic;

      float t = since/dur;
      if (d < 2.0*t) {
        if (t < sqrt2) { // color in
          react = vec4(clr);
        } else if (t < pi) { // fade out
          float fac = 1.0 - (t-sqrt2)/(pi-sqrt2);
          react = vec4(fac*clr);
        }
      }

      gl_FragColor += react;
    } else {
      discard;
    }
	}
}
