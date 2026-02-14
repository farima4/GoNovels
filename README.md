# **GoNovels** - barebones novel hosting webserver

## **PROJECT IS STILL NOT FUNCTIONAL**

well it technically is but not really

This project is not meant to be a fully flashed out novel hosting service with acounts.

The objective here is to provide a simple way to host your personal novel collection. Be it (ideally) your own, self-written, web-novels or perhaps a collection of some other novels (if you care enough to port it to markdown).

I got the idea to do it because I like writing funny things for fun sometimes and I didn't really wish to use other services (like Wattpad, ew) to upload my works.

if you care enough to read the slop I have written (or more likely you just want to see how the site looks/functions) it is (should be at least) available at:

[novels.farima4.space](https://novels.farima4.space)

## Rough explanation of how it works:

This project doesn't use a database.

In the root directory of the project is a `novels/` folder. inside of it you are supposed to place your novels, each saparated in to a different folder.

- (note: The `novels/` folder for now has to stay in the root directory, I might add the ability to point to a different path later, for example `~/Documents/novels`)

- (note: I will provide an example novel with the project by default to understand it easier)

How you name the folder doesn't technically matter but you should keep the folder names the same as the novel names for your own sake.

So create a folder `novels/my-first-novel`.

inside the `my-first-novel` folder the most important file to create is `metadata.json`

the `novels/my-first-novel/metadata.json` should look like this:

```json
{
  "title": "Epic Novel Title!",
  "description": "An amazing and thrilling description that will get everyone hooked immediately!",
  "author": "farima4",
  "cover": "cool picture.jpg"
}
```

while `title` and `description` are self explanitory, `cover` might be a bit more elusive.

your novel can also have the `media/` subfolder, and `cover` points to a picture exclusively inside of the media folder. So the entire path would be `"novels/my-first-novel/media/cool picture.jpg"`

- (note: if not cover is provided or if it doesn't exist, it will fallback to a placeholder in `static/cover.jpg`)
- (note: I will add more info about file extentions later)

So now we have all the information, we just need to add the chapters and that is quite easy. Chapters are markdown files (.md) and they are placed directly in to the root directory of that novel besides the `metadata.json` (so directly in `novels/my-first-novel`). only 2 rules must be followed:

1. Chapters must follow the following naming convention: `chapter-n.md`, where the `n` should be replaced by the chapter number.

2. The first line of the chapter must look like this: `# [This is the title of the chapter]`. This is so I can grab the title directly from the file and display it on the website

So the `novels` folder should look like this:

```bash
my-first-novel/
├── chapter-1.md
├── chapter-2.md
├── media
│   └── 'cool picture.png'
└── metadata.json
```

While `chapter-1.md` should look like

```markdown
# A new morning, a new beginning!

Farima4's eyelids quivered as the morning sunshine washed accross his sleeping figure.
. . .
```

After this you are free to use markdown however you wish, even add tables to your novels if you so wish.

This is also why I have the `media/` folder, so you can freely upload pictures and/or even videos and then link them directly in the markdown file using the following syntax `![Some Alt Text](media/my-world-map-or-something.jpg)` and it will be displayed in the chapter!

**will write the rest of readme later**
Credits to (GentooPegin)[https://github.com/GentooPegin] for making the html templates
