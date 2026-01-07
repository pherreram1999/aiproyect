import sqlite3
con = sqlite3.connect("score.db")
cur = con.cursor()
cur.execute("SELECT name FROM sqlite_master WHERE type='table';")
tables = [r[0] for r in cur.fetchall()]
print("Tables:", tables)
for t in tables:
    print(f"\n-- {t} --")
    for row in cur.execute(f"SELECT * FROM {t} LIMIT 10"):
        print(row)
con.close()