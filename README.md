# Aguaxaca 🚰

As of 2025, Oaxaca's water distribution schedules are shared daily
through images, on two social networks: Facebook and X (formerly
Twitter). I imagine the image format is a workaround to message-length
and formatting limitations of these platforms.

This is annoying for three reasons:

1. storing this data on centralized US silos isn't good for freedom:
   you're forced to create accounts to be informed, and they get to keep
   all the public data, and
2. it excludes people with visual disabilities, and
3. it makes hard to archive, or analyze the data.

This program won't solve these issues because they are not entirely technical,
but it's a small workaround to:

1. extract data from public notices when they are published,
2. share the same data in text form (more accessible!), and
3. provide access to historical records (analytics!).

> [!CAUTION]
> 🚧 The project is a work-in-progress, evolving when free time allows it.
> Don't rely on it too much.

A public instance is running at https://agua.cypr.io

# Self hosting

## Using docker

A free docker image is published on Docker Hub.

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
See below for a list of known variables, for example.

```
docker run ozzz/aguaxaca:latest \
  -v ./data:/data \
  -e /etc/default/aguaxaca \
  aguaxaca collect
```

## Environment variables

**Required**, for the `analyze` sub-command:

- `OPENROUTER_API_KEY`: OpenRouter API key, to extract text from images.

**Required**, for the `collect` sub-command:

- `TWITTERAPI_IO_API_KEY`: API key for [twitterapi.io](https://twitterapi.io),
  used to fetch the latest tweets and their images.

**Optional**, for the `collect` sub-command:

- `COLLECTOR`: which backend to use: `twitterapi` (default) or `nitter`.
- `NITTER_HOST`: where we fetch tweets when using the `nitter` backend,
  defaults to `https://nitter.net` (deprecated).
- `TWITTER_ACCOUNT`: Twitter/X handle, defaults to `SOAPA_Oax`.

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

By default we collect data with [twitterapi.io](https://twitterapi.io),
a paid REST API for Twitter/X. We only fetch the latest tweets from
the `@SOAPA_Oax` account, to download the attached photos with delivery
schedules. Based on the current publication schedule, running collection
about 3-4 times per day is enough, and costs a fraction of a cent per
run.

You must set `TWITTERAPI_IO_API_KEY` to your API key.

### Nitter

The project initially relied on _Nitter_, a free and open-source project
to scrape data from Twitter/X. This was not very reliable and is now
deprecated, but you can still use it!

Set `COLLECTOR=nitter` to scrape tweets from a
[Nitter](https://nitter.net) instance instead (⚠️ Nitter is
unmaintained and public instances are mostly gone).

With the Nitter collector backend, set the `NITTER_HOST` env. variable
to point to your own private instance (scraping other public instances
without permission is rude).

## Image analysis

To extract text, and store data from downloaded images, run:

```
aguaxaca analyze
```

Tesseract is an okay open-source OCR program, but because of the layout
of SOAPA's images, it won't work well here. Instead, we rely on AI tools
for text extraction, and formatting the output.

We use OpenRouter's API (with Thinking Machines's inkling-2-small model
for now) to extract information from the images. This means that you
will need an OpenRouter API key with some credits to run the parser.

Currently, this costs a few cents per image.

With the correct hardware, using a local model would also work, but
that's more expensive than OpenRouter for now, because big AI vendors
like to make RAM and GPUs very expensive for the rest of us. 💸

Look into `parser/parser.go` for a prompt that will extract information from
SOAPA_Oax's publications. Here's a sample response:

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
