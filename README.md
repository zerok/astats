# AStats

astats is a simple access-log based statistics reporter for your website. The
main goal of this is to just "ingest" logfiles and store aggregated results
into a SQLite database. After that, the logfile can be deleted. The aggregate
does *not* contain any identifyable information but just things like page-views
per day and referrers for each URL (not yet implemented).

**Note:** At this point astats only works with Caddy 2 JSON log files!


## Usage

```
# Ingest a logfile into astats.sqlite
$ astats ingest /srv/www/yoursite.com/logs/access.json.log

# Show the top pages for today
$ astats query top
```
