# URL Shortener

Super simple URL shortener that stores everything in memory and cleans itself up every hour. No DB, no fuss.

POST your long URL to `/shorten` like:

```json
{
  "long_url": "https://your-super-long-url.com"
}
```
