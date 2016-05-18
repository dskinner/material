# material

This package is a work-in-progress for providing an implementation of material design for gomobile and potentially exp/shiny.

The core goal of this package is to provide an implementation of material design as defined here: https://www.google.com/design/spec/material-design/introduction.html

Nothing more.

Features currently provided were written to determine the exact nature of moving forward with a proper implementation. These include the following:

* Constraint based layouts using simplex method. See https://github.com/dskinner/simplex package. A trivial box model has been written on top of this. The intended usage of the constraint based layout is to fulfill the adaptive layout requirements of material design.

* Key light shadows. These shadows are not approximated on the material. This can be seen by tilting a view matrix and inspecting. The current implementation is fairly performant on mid ranged android devices.

* Material design icons, mdpi. See material/icon package. This currently provides 913 icons at 48x48 px as a single texture. While not suitable for hires display, this is more than suitable for development of the package itself. A script in material/icon can be expanded to provide desired sets of icons at high resolution. Grab a copy of the texture here: https://drive.google.com/a/dasa.cc/file/d/0B6hxg-gC2Uz_cG1DakFNcDFxYlk/view

* Text is provided with signed-distance-field texture. See material/text source for generating a texture and material/example/text for usage.

* Material. This also includes 1dp of thickness, though the exact nature of this fades when transforming rectangles to ellipses. Only two behavior flags are currently recognized declaring material flat (transparent) or raised (has shadow).

* Color. A full list of colors as defined is the spec is available as uint32s, e.g. BlueGrey500

* Environment. While more of an abstraction in the design spec, there is an environment type that accepts a theme palette and is used for creation of new material types such as button and toolbar.

## Roadmap

One of the most important aspects of any UI is to have an effective way to layout items. Such a solution is purely numerical. With that in mind, the current focus is to provide an implementation of material design's [Adaptive UI](https://www.google.com/design/spec/layout/adaptive-ui.html). The implementation should not be interdependent on any other portion of package material. It's conceivable for example to make use of such a package for reflowing terminal based UIs.

The package shall be well developed, documented, and tested. There are enough minimal implementations of other portions of material design to provide meaningful examples of such a package once complete. Only upon completion of Adaptive UI will a determination of the next step be made.

As for a timeframe on the whole of package material with only the spare time of a single developer, I could only wildly guesstimate a time frame of 2+ years for completion. Part of this is a strong urge to take a slow-and-steady pace for each portion of package material to best ensure design choices, proper documentation, and reasonably complete testing.

## Contributing

Everything in this package is in-flux. If you're interested in contributing, understand the current focus is iterating on the currently provided features. New features will not be accepted unless they are important for establishing a baseline in determining future functionality and api. Anything beyond the material design spec is out of scope for this package and will not be accepted.

Please open an issue and discuss your thoughts first.
