# Aguaxaxa

As of 2025, Oaxaca's water distribution schedules are shared daily
through images, on two social networks: Facebook and X (formerly
Twitter). I guess it's a workaround to message length and format
limitations on these platforms, but it also excludes people with visual
disabilities, and makes it impossible to search through this data.

This also forces the citizens to create accounts on privately owned
social networks. There is no public historical data about the schedules.

This program won't solve these issues because they're not all technical,
but it's a workaround to:

  1. extract data from public notices,
  2. share the same data in text form (accessible!), and
  3. providing access to some historical records.

**The project is still very much a work-in-progress. Don't use it now!**

# Technical information

The project is organized as follow:

- `app/` —  the core application types.
- `collector/` —  collect images from social networks.
- `parser/` —  parse collected images into structured data structures.

## Data collection

Data collection currently relies on [Nitter](https://nitter.net), a pure
HTML front-end for X, to fetch the images from the @SOAPA_Oax account.
We rely on the main instance of Nitter **for now**, with a scraper to get
the data once or twice a day, because that's our source's publication
schedule.

This means about 3-4 requests per 24h, which doesn't seem abusive for
testing / developing, **but** run your own Nitter instance, or get
permission to use a public instance, for a long-term solution.

Other improvements to explore:

- Scrape X directly (probably stupidly expensive these days), or
- scrape Facebook, or
- use Nitter's RSS features instead of scraping (RSS is not always
  available on public Nitter instances).

### Dev notes

- [x] if an image is already on the local cache do not "return" it.
- [ ] cleanup: we'll want to delete images after a couple of weeks or
      so. We could get by with 24h cache, but let's play it safe in case
      we want to debug stuff.

## Text parsing

Tesseract has always been an okay open-source OCR program, but because
of the image formatting, it is currently unable to extract text from the
notices. To work around this, we currently rely on genAI for OCR-ing
*and* formatting the output.

The code is tailored for Anthropic's LLM (Sonnet 3.7 is completely fine)
to extract information from the images. This means that you will need an
Anthropic API key, and some credits, to run the "parser". Yay.

With the correct hardware, using Ollama with a different model would
work great for example, but that's more expensive than paying Anthropic
for now.

Look into `parser/parser.go` for a prompt that will extract information from
SOAPA_Oax's publications. Here's a sample response from Sonnet 3.7:

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

### Dev Notes

- [x] Build a go client to query https://docs.anthropic.com/en/api/messages
- [ ] CSV parser for LLM response, to a nicer data structure for storage.

## Data store


### Dev notes

No need for anything very fancy. Sqlite will be fine for a good while.

It would be nice to search and match names like "América" if we type
"amér" or even "ame": ignoring case, and accentuated characters, buy
we'll need more than Sqlite at that point.

## Web

1. Use a small router (e.g. Chi) to serve the data as JSON.
2. Build a light front-end, mobile first.
3. Cache all the things.

Some ideas for the routes.

### Search distribution schedules

`GET /schedules`

Without filter: get the latest schedules in the DB.

Limit to the last 3-4 days by default.

`GET /schedules?since=YYYYMMDD`

Get all schedules since date. This should probably have a limit, with a
notice to contact to get access to the full dataset.

`GET /schedules/ids[]=1`

Get schedules for one or several zone IDs (supposing that we store zones
in their own table).

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
