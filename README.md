# Aguaxaxa

As of 2025, Oaxaca's water distribution schedules are shared daily
through images, on two social networks: Facebook and X (formerly
Twitter). This forces the inhabitants to create accounts on these social
networks to know when they may receive the precious liquid. There is no
public historical data about the schedules, and the information published as
raw images isn't accessible.

This program provides a solution for these issues by:

  1. extracting data from public notices,
  2. sharing the same data in text form, and
  3. providing access to previous distribution dates.

**The project is still very much a work-in-progress. Don't use it now!**

# Technical information

The project is organized as follow:

- `app/` —  the core application type.
- `collector/` —  collect images from social networks.
- `parser/` —  parse collected images into structured data structures.

## Data collection

Data collection currently relies on [Nitter](https://nitter.net), a pure
HTML front-end for X, to fetch the images from the @SOAPA_Oax account.
We rely on the main instance of Nitter, with a scraper to get the data
once or twice a day.

**This should mean about 3-4 requests per 24h, which doesn't seem
abusive but running your own Nitter instance is a good idea.**

This could be improved in a few ways. For example:

- Scrape X directly, or
- Host our own private Nitter instance,
- Use Nitter's RSS features instead of scraping (RSS is not always
  available on public Nitter instances).

### Dev notes

- [x] if an image is already on the local cache do not "return" it.
- [ ] cleanup: we'll want to delete images after a couple of weeks or
      so. We could get by with 24h cache, but let's play it safe in case
      we want to debug stuff.

## Text parsing

Tesseract is a good open-source OCR program, but because of the image
formatting, it is currently unable to extract text from the notices. To
work around this, we currently rely on Anthropic's LLM (Sonnet 3.7) to
extract information from the images.

This means that you will need an Anthropic API key, and some credits, to
run the "parser". Yay.

Here's a prompt that will successfully extract data from the water
distribution images:

See `parser/parser.go` for a prompt that will extract information from
SOAPA_Oax's publications.

Here's a sample answer from Sonnet 3.7:

```
date,schedule,location_type,location_name
2025-03-16,matutino-vespertino,COLONIA,Unión
2025-03-16,matutino-vespertino,COLONIA,Periodista
2025-03-16,matutino-vespertino,COLONIA,Francisco I. Madero
2025-03-16,matutino-vespertino,COLONIA,Ixcotel
2025-03-16,matutino-vespertino,COLONIA,Carretera antigua a Monte Albán
2025-03-16,matutino-vespertino,COLONIA,Hospital Civil
2025-03-16,matutino-vespertino,FRACCIONAMIENTO,Elsa
2025-03-16,matutino-vespertino,EJIDO,Guadalupe Victoria (sectores 3 y 4 parte baja)
2025-03-16,vespertino-nocturno,COLONIA,Reforma (parte baja)
2025-03-16,vespertino-nocturno,COLONIA,América Norte (parte alta)
2025-03-16,vespertino-nocturno,COLONIA,Lomas del Crestón (parte baja y media)
2025-03-16,vespertino-nocturno,COLONIA,Estrella (parte baja)
2025-03-16,vespertino-nocturno,COLONIA,Eliseo Jiménez Ruiz Norte
```

### Dev Notes

- [x] Build a go client to query https://docs.anthropic.com/en/api/messages
- [ ] CSV parser for LLM response, to a nicer data structure for storage.

## Data store


### Dev notes

No need for anything very fancy. Sqlite will be fine for a good while.

It would be nice to search and match names like "América" if we type
"amér" or even "ame": ignoring case, and accentuated characters.

## Web

1. Use a small router (e.g. Chi) to serve the data as JSON.
2. Build a light front-end, mobile first.

### Search distribution schedules

`GET /schedules`

Without filter: get the latest schedules in the DB.

Limit to the last 24h by default.

`GET /schedules?since=YYYYMMDD`

Get all schedules since date.

`GET /schedules/ids[]=1`

Get schedules for one or several zone IDs.

`GET /schedules?zone=%s`

Get the schedules in the DB, for zones that match the search string.

### Find distribution points

`GET /zones?q=%s`

List all known zones, or zones matching a string.

We should show the some basic stats:

- number of days since previous water distribution,
- the average number of days between each distributions,
- probable number of days until next distribution.

The distribution schedules aren't regular: they will vary with draught
and water levels. With enough data, we may be able to predict the
distribution data for a specific point.
