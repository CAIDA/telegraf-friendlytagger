package main

import (
    "fmt"
    "log"
    "os"
    "time"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "context"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type PandaAs struct {
    Id          int32
    Score       int32
    Asn         string
    AsnName     string
    Country     string
    Source      string
    OrgId       string
    OpaqueId    string
    Changed     string
    Date        string
    Ts          string
}

type PandaPageInfo struct {
    PageSize    int32
    PageOffset  int32
    HasNextPage bool
}

type PandaJsonBlob struct {
    TotalCount      int32
    PageInfo        PandaPageInfo
    Errors          string
    Data            []PandaAs
}


func LoadAsnNames(pageid int32, db *sql.DB) bool {

    var ctx context.Context

    url := fmt.Sprintf("https://api.panda.caida.org/as2org/dev/asns/?verbose=true&page=%d", pageid)

    client := http.Client{ Timeout: time.Second * 10}

    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        log.Fatal(err)
    }

    req.Header.Set("User-Agent", "telegraf-friendlytagger")
    res, getErr := client.Do(req)
    if getErr != nil {
        log.Fatal(getErr)
    }

    if res.Body != nil {
        defer res.Body.Close()
    }

    body, readErr := ioutil.ReadAll(res.Body)
    if readErr != nil {
        log.Fatal(readErr)
    }

    jsonBlob := PandaJsonBlob{}
    jsonErr := json.Unmarshal(body, &jsonBlob)
    if jsonErr != nil {
        log.Fatal(jsonErr)
    }

    ctx = context.Background()
    tx, txErr := db.BeginTx(ctx, &sql.TxOptions{Isolation:sql.LevelSerializable})
    if txErr != nil {
        log.Fatal(err)
    }

    for _, asn := range(jsonBlob.Data) {
        if asn.AsnName == "" {
            asn.AsnName = "Name Unknown"
        }
        t, timeErr := time.Parse(time.RFC3339, asn.Changed)
        if timeErr != nil {
            log.Fatal(timeErr)
        }

        query := "INSERT OR IGNORE INTO asn_mappings(code, label, apply_from) VALUES (?, ?, ?)"
        combinedName := fmt.Sprintf("%s, %s", asn.AsnName, asn.Country)

        _, queryErr := tx.Exec(query, asn.Asn, combinedName, t.Unix())
        if queryErr != nil {
            _ = tx.Rollback()
            log.Fatal(queryErr);
        }
        fmt.Printf("%s %s, %s -- %d\n", asn.Asn, asn.AsnName, asn.Country,
                t.Unix())
    }

    if commitErr := tx.Commit(); commitErr != nil {
        log.Fatal(commitErr)
    }

    return jsonBlob.PageInfo.HasNextPage
}

func main() {

    db, err := sql.Open("sqlite3", os.Args[1])
    if err != nil {
        log.Fatal(err)
    }

    tabcreate := "CREATE TABLE IF NOT EXISTS asn_mappings (code text NOT NULL, label text NOT NULL, apply_from INTEGER, apply_to INTEGER, UNIQUE(code, apply_from))"

    _, err = db.Exec(tabcreate)
    if err != nil {
        log.Fatal(err)
    }

    index := int32(1)
    more := LoadAsnNames(index, db)

    for more == true {
        index = index + 1
        more = LoadAsnNames(index, db)
    }

    db.Close()
}

// vim: set sw=4 tabstop=4 softtabstop=4 expandtab :
