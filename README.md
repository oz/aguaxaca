# Aguaxaca 🚰

As of 2025, Oaxaca's water distribution schedules are shared daily through
images, on two social networks: Facebook and X (formerly Twitter). I imagine
that it's a workaround to message-length and formatting limitations of these
platforms. This is still annoying. It excludes people with visual disabilities,
and also makes it hard to use, or search, the data. This also forces the
citizens to create accounts on privately owned social networks.

This program won't solve these issues because they are not entirely technical,
but it's a small workaround to:

1. extract data from public notices as they are published,
2. share the same data in text form (accessible!), and
3. provide access to historical records.

> [!CAUTION]
> 🚧 The project is a work-in-progress, evolving when free time allows it.
> Don't rely on it too much.

A public instance is running at https://agua.cypr.io

# Self hosting

## Using docker

A docker image is published on Docker Hub.

To run the web server locally:

```
docker run ozzz/aguaxaca:latest -p 8080:8080 -v ./data:/data
```

The `./data` directory is where the program will store collected image
files, and its SQLite database.

Periodically, you can fetch fresh data by running:

```
docker run ozzz/aguaxaca:latest -v ./data:/data aguaxaca collect
```

To parse the data, usually right after a successfull "collect" run, use:

```
docker run ozzz/aguaxaca:latest -v ./data:/data aguaxaca analyze
```

Use `-e` or `--env-file` to pass relevant env. variables to the container.
See below for a list of known variables.

## Environment variables

Required, for the *analyze* sub-command:

- `ANTHROPIC_API_KEY`: Anthropic private API key, to extract text from images.

Optional, for the *collect* sub-command:

- `NITTER_HOST`: where we fetch tweets, defaults to `http://nitter`.
- `NITTER_ACCOUNT`: Twitter/X handle, defaults to `SOAPA_Oax`.

# Technical information

The repository is organized as follow:

- `app/` —  the core application types.
- `collector/` —  collect images from social networks.
- `parser/` —  parse collected images into structured data structures.
- `web/` —  web server.

## Data collection

To download public schedules shared recently, run:

```
aguaxaca collect
```

Data collection currently relies on [Nitter](https://nitter.net), a pure HTML
front-end for X, to fetch the images from the @SOAPA_Oax account. Based on the
current publication schedule, running colleciton about 3-4 times per day is
enough.

Running your own private Nitter instance is not much work. To use a public
instance change the `NITTER_HOST` environment variable, and ask for permission maybe.

Other improvements to explore:

- Scrape X directly (probably stupidly expensive these days), or
- scrape Facebook (really?), or
- use Nitter's RSS feed instead of scraping (though RSS is often disabled on
  public Nitter instances).

## Image analysis

To extract text, and store data from downloaded images, run:

```
aguaxaca analyze
```

Tesseract is an okay open-source OCR program, but because of the layout of
SOAPA's images, it won't work well here. Instead, we rely on genAI for OCR-ing
*and* formatting the output.

We use Anthropic's APIs to extract information from the images (target is Sonnet
4 model for now). This means that you will need an Anthropic API key with some
credits to run the parser. Currently, this costs a few cents per image.

With the correct hardware, using a local model would also work, but that's way
more expensive than Anthropic for now. 💸

Look into `parser/parser.go` for a prompt that will extract information from
SOAPA_Oax's publications. Here's a sample response from Sonnet 4.0:

```
date,schedule,location_type,location_name
2025-07-21,matutino-vespertino,colonia,Libertad
2025-07-21,matutino-vespertino,colonia,10 de Abril
2025-07-21,matutino-vespertino,colonia,Monte Albán
2025-07-21,matutino-vespertino,colonia,Adolfo López Mateos
2025-07-21,matutino-vespertino,colonia,Presidente Juárez
2025-07-21,matutino-vespertino,colonia,Margarita Maza de Juárez
2025-07-21,matutino-vespertino,colonia,Bosque Sur
2025-07-21,matutino-vespertino,colonia,"Jardín (sector Bugambilias)"
2025-07-21,matutino-vespertino,fraccionamiento,Jardines de Las Lomas
2025-07-21,matutino-vespertino,ejido,"Guadalupe Victoria (sector 1, 2ª sección Oeste)"
2025-07-21,matutino-vespertino,unidad,Ferrocarrilera
```

## Data store

### Dev notes

Sqlite should be fine for a long while, with FTS5 providing full-text search on locations.

As we get more data, we could provide more services:

1. figure out unique IDs for each zone —  the original data, with district names, is often incoherent and not precise.
2. compute some stats like: delivery interval in days, number of deliveries tracked per year, etc.
3. provide an export function for people interested in the raw data.
