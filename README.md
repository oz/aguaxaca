# Aguaxaca

As of 2025, Oaxaca's water distribution schedules are shared daily
through images, on two social networks: Facebook and X (formerly
Twitter). I guess it's a workaround to format and message length
limitations of these platforms, but it excludes people with visual
disabilities, and also makes it impossible to search through this data.

This also forces the citizens to create accounts on privately owned
social networks. There is no public historical data about the schedules,
etc.

This program won't solve these issues because they are not all
technical, but it's a small workaround to:

1. extract data from public notices,
2. share the same data in text form (accessible!), and
3. providing access to some historical records.

**The project is still very much a work-in-progress. Don't use it now!**

# Technical information

The repository is organized as follow:

- `app/` —  the core application types.
- `collector/` —  collect images from social networks.
- `parser/` —  parse collected images into structured data structures.

## Data collection

To download public schedules shared recently, run:

```
aguaxaca collect
```

Data collection currently relies on [Nitter](https://nitter.net), a pure
HTML front-end for X, to fetch the images from the @SOAPA_Oax account.
We rely on the main instance of Nitter **for now**, with a scraper to get
the data once or twice a day, because that's our source's publication
schedule.

This means about 3-4 requests per 24h, which doesn't seem abusive for
testing / developing, **but** for a long-term solution: run your own
Nitter instance, or get permission to use a public instance. Be nice,
okay? 😊

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

To extract text, and store data from downloaded images, run:

```
aguaxaca analyze
```

Tesseract is an okay open-source OCR program, but because of the layout
of SOAPA's images, it won't work well here. Instead, we rely on genAI
for OCR-ing *and* formatting the output.

The code is tailored for Anthropic's APIs to extract information from
the images (target is Sonnet 3.7 model for now). This means that you
will need an Anthropic API key, and some credits, to run the parser.
Yay. With the correct hardware, using a local model would also work, but
that's more expensive than Anthropic for now. 💸

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

- [x] Go client/lib to query https://docs.anthropic.com/en/api/messages
- [x] CSV parser for LLM response, to a nicer data structure for storage.
- [ ] Add CLI flag to store LLM responses on disk.


## Data store

### Dev notes

Sqlite should be fine for a long while.

It would be nice to search and match names like "América" if we type
"amér" or even "ame": ignoring case, and accentuated characters, but
we could need more than the basic Sqlite at that point. See below:

- [ ] Check Sqlite's "FTS" module, and create/update triggers.
- [ ] Backup with litestream.

## Web

1. Use a small router (e.g. Chi?) to serve the data as JSON, or
   plain HTML, with some HTMX JS sprinkled on top, or both...
2. Build a *light* front-end anyway, mobile first.
3. Cache all the things.
4. Track stuff with a prometheus endpoint, for fun.

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
