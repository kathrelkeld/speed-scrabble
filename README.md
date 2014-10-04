# Speed Scrabble

## Sass installation

Install [sassc](https://github.com/sass/sassc). Quick version:

```
cd /tmp
git clone https://github.com/sass/sassc.git
git clone https://github.com/sass/libsass.git
SASS_LIBSASS_PATH=libsass make
cp bin/sassc ~/bin/ # assuming you have ~/bin and it's in your path
```

You can use the following [Reflex](https://github.com/cespare/reflex) command to watch and rebuild the CSS:

    reflex -g css/*.scss make
