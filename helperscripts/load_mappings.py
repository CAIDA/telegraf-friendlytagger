#!/usr/bin/python3

import sys, string, csv, sqlite3

try:
        dbname = sys.argv[1]
        srcfile = sys.argv[2]
        table = sys.argv[3]
except:
        print("Usage: %s <dbname> <source file> <table name>" % (sys.argv[0]))
        sys.exit(-1)

conn = sqlite3.connect(dbname)
c = conn.cursor()

c.execute("CREATE TABLE IF NOT EXISTS %s (code text not null, label text not null, apply_from integer, apply_to integer)" % (table))
conn.commit()

with open(srcfile, newline='') as csvfile:
        rdr = csv.reader(csvfile)
        for row in rdr:
                if row[2] == "0":
                        row[2] = None
                if row[3] == "0":
                        row[3] = None
                stmt = "INSERT into %s values (?, ?, ?, ?)" % (table)
                c.execute(stmt, (row[1], row[0], row[2], row[3]))

                print(row[0])
        conn.commit()
conn.close()

