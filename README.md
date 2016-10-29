## Simple CMS & API written in Go

This is a CMS & API written in Go. I've never written such a system before, so Go has wonderful opportunity – this ecosystem encourages usage of small libraries, not big frameworks.

So far I have next ecosystem:

- [gin](github.com/gin-gonic/gin) framework for routing and static serving
- [sql driver](github.com/go-sql-driver/mysql)
- [imgio](github.com/anthonynsimon/bild/imgio) for image processing

For the client-side (js and styles) I have a gulp (due to large ecosystem and absence of babel). `Sass` for the styles, and plain ES5 javascipt. The latter is kind of boring, but babel is too big nowadays, and until I have bunch of modules to share, I don't really want to introduce it here. For any other language, it won't work, because it is all about 3rd-party libraries, so best interoperability is must, and nothing has it (except coffeescript, but who cares nowadays about it).
