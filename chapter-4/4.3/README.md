%date

2024-07-29

**Important**: It’s crucial to point out here that all the fields in our Movie struct are exported (i.e. start with a capital letter), which is necessary for them to be visible to Go’s encoding/json package. Any fields which aren’t exported won’t be included when encoding a struct to JSON

```go
type Movie struct {
    ID int64
    CreatedAt time.Time
    Title string
    Year int32
    Runtime int32
    Genres []string
    Version int32
}
```

NOTE: jq command to pretty print the json on bash

```bash
$ sudo apt install jq
$ curl  localhost:4000/v1/movies/123 | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   152  100   152    0     0   131k      0 --:--:-- --:--:-- --:--:--  148k
{
  "ID": 123,
  "CreatedAt": "2024-07-29T10:55:38.905889076+05:30",
  "Title": "Casablanca",
  "Year": 0,
  "Runtime": 102,
  "Geners": [
    "drama",
    "romance",
    "war"
  ],
  "Version": 1
}
```

## Struct tags

```go
type Movie struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    Title     string    `json:"title"`
    Year      int32     `json:"year"`
    Runtime   int32     `json:"runtime"`
    Genres    []string  `json:"genres"`
    Version   int32     `json:"version"`
}
```

```bash
{
  "id": 123,
  "created_at": "2024-07-29T11:18:24.784814493+05:30",
  "title": "Casablanca",
  "year": 0,
  "runttime": 102,
  "genres": [
    "drama",
    "romance",
    "war"
  ],
  "version": 1
}
```

## "omitempty" struct tag
</br>
In contrast the omitempty directive hides a field in the JSON output if and only if the struct field value is empty, where empty is defined as being:

1. Equal to false, 0, or "".
2. An empty array, slice or map.
3. A nil pointer or a nil interface value

**Note**: You can also prevent a struct field from appearing in the JSON output by simply making it unexported. But using the json:"-" struct tag is generally a better choice: it’s an explicit indication to both Go and any future readers of your code that you don’t want the field included in the JSON, and it helps prevents problems if someone changes the field to be exported in the future without realizing the consequences.


### Wrong way ("with space")
```go
type Movie struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"-"` // Use the - directive
    Title     string    `json:"title"`
    Year      int32     `json:"year, omitempty"`    // Notice the <space> before omitempty. This <space> shouldn't be there.
    Runtime   int32     `json:"runtime, omitempty"` // Add the omitempty directive
    Genres    []string  `json:"genres, omitempty"`  // Add the omitempty directive
    Version   int32     `json:"version"`
}
```

```bash

 curl  localhost:4000/v1/movies/123 | jq 
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   103  100   103    0     0  91637      0 --:--:-- --:--:-- --:--:--  100k
{
  "id": 123,
  "title": "Casablanca",
  "year": 0,
  "runttime": 102,
  "genres": [
    "drama",
    "romance",
    "war"
  ],
  "version": 1
}
```

### Correct way 
```go
type Movie struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"-"` // Use the - directive
    Title     string    `json:"title"`
    Year      int32     `json:"year,omitempty"`  
    Runtime   int32     `json:"runtime,omitempty"`
    Genres    []string  `json:"genres,omitempty"`
    Version   int32     `json:"version"`
}
```

```bash
statemate@statemate:~/work/gostuff/greenlight/chapter-3/3.3$ curl  localhost:4000/v1/movies/123 | jq 
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    94  100    94    0     0  92066      0 --:--:-- --:--:-- --:--:-- 94000
# Notice that year field didn't show up as it had 0 value.
{
  "id": 123,
  "title": "Casablanca",
  "runttime": 102,
  "genres": [
    "drama",
    "romance",
    "war"
  ],
  "version": 1
}
```

**Hint**: If you want to use omitempty and not change the key name then you can leave it blank in the struct tag — like this: json:",omitempty". Notice that the leading comma is still required.

### The string struct tag directive

A final, less-frequently-used, struct tag directive is string. You can use this on individual struct fields to force the data to be represented as a string in the JSON output.

For example, if we wanted the value of our Runtime field to be represented as a JSON string (instead of a number) we could use the string directive like this:


```go
type Movie struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"-"` // Use the - directive
    Title     string    `json:"title"`
    Year      int32     `json:"year,omitempty"`  
    Runtime   int32     `json:"runtime,omitempty,string"`
    Genres    []string  `json:"genres,omitempty"`
    Version   int32     `json:"version"`
}
```

```bash
{
  "id": 123,
  "title": "Casablanca",
  "runttime": "102", // <-- This is string now
  "genres": [
    "drama",
    "romance",
    "war"
  ],
  "version": 1
}
```

**Note** that the string directive will only work on struct fields which have int*, uint*, float* or bool types. For any other type of struct field it will have no effect.

## json.MarshalIndent()

MarshalIndent() function will automatically adds whitespaces to the JSON output, putting each element on a separate line and prefixing each line wiht optional prefix and indent characters.


In these benchmarks we can see that json.MarshalIndent() takes 65% longer to run and uses around 30% more memory than json.Marshal(), as well as making two more heap allocations. Those figures will change depending on what you’re encoding, but in my experience they’re fairly indicative of the performance impact.
</br>
If your API is operating in a very resource-constrained environment, or needs to manage extremely high levels of traffic, then this is worth being aware of and you may prefer to stick with using json.Marshal() instead.

## Custom MarshallJSON()

```go
type Marshaler interface {
    MarshalJSON() ([]byte, error)
}
```

```json
{
  "movie": {
    "id": 123,
    "title": "Casablanca",
    "runttime": "102 mins",
    "genres": [
      "drama",
      "romance",
      "war"
    ],
    "version": 1
  }
}
```

## Curl command to upload the files
```bash
 curl -d '{"title": "Moana"}{"title": "Top Gun"}' localhost:4000/v1/movies
 curl -d '{"title": "moana"}' localhost:4000/v1/movies
 curl -d '{"title": "moana", "rating":"PG"}' localhost:4000/v1/movies
 curl -d @/tmp/largefile.json localhost:4000/v1/movies

```

