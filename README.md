# Open Gacha Codes

Open Gacha Codes is an API service that collects active redemption codes for supported gacha games from multiple independent sources.

It normalizes and deduplicates codes, combines reward information, excludes unsupported regional codes, and stores canonical results for API consumers.

## Routes

### `GET /games`

Returns all supported games.

```json
[
  {
    "slug": "genshin",
    "name": "Genshin Impact"
  },
  {
    "slug": "starrail",
    "name": "Honkai: Star Rail"
  },
  {
    "slug": "zenless",
    "name": "Zenless Zone Zero"
  },
  {
    "slug": "endfield",
    "name": "Arknights: Endfield"
  }
]
```

### `GET /games/{slug}/codes`

Returns active redemption codes for a supported game using a slug returned by `GET /games`.

```json
[
  {
    "code": "ZENLESSGIFT",
    "rewards": [
      "Polychrome x50",
      "Official Investigator Log x2",
      "W-Engine Power Supply x3",
      "Bangboo Algorithm Module x1"
    ]
  }
]
```

A supported game with no active codes returns an empty array:

```json
[]
```

## Error Responses

Errors use a consistent response structure. For example, an unknown game returns HTTP `404` with:

```json
{
  "statusCode": 404,
  "error": {
    "message": "game not found"
  }
}
```

Other possible error statuses include `405` for unsupported HTTP methods and `500` for internal server errors.

## Request a Game

If you would like another game added to the tracker, feel free to open an issue. When possible, include links to reliable sources that publish the game's active redemption codes, such as an official website, community wiki, or regularly maintained guide.
