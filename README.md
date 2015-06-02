# BusProxy
Streamline the data from Avail's feed into a single request

## The API

It's super simple. Just hit the root URL ("/") with query parameters containing stopIDs, like so:

`GET /?stop=64&stop=63`

And the API endpoint will return a JSON formatted object comprised of an array of `StopResponse` objects, which have the following structure:

```Go
type StopResponse struct {
	StopName string      `json:"stopName"`
	Routes   []RouteInfo `json:"routes"`
}

type RouteInfo struct {
	Number        string          `json:"number"`
	Name          string          `json:"name"`
	DepartureTime avail.AvailTime `json:"departureTime"`
	Color         string          `json:"color"`
	TextColor     string          `json:"textColor"`
}
```
