#!/bin/env python3
import requests
import mysql.connector
import datetime
import sys
import zipfile
import os
import shutil
from config import config

db = mysql.connector.connect(
        host=config['DB_HOSTNAME'],
        user=config['DB_USER'],
        passwd=config['DB_PASSWORD'],
        database=config['DB_TMP']
    )
curs = db.cursor(dictionary=True)



def getLastUpdate():
    """
    Check in the database the date of the last update
    """
    curs.execute("""SELECT lastUpdate
                    FROM TEC.meta
                    WHERE Id = 1""")
    lastUpdate = curs.fetchone()
    if lastUpdate is not None:
        lastUpdate = lastUpdate['lastUpdate']
    return lastUpdate

def getLastGTFS(url):
    """
    Gets the date of the last modification on the GTFS source zip
    """
    req = requests.head(url)
    if req.ok == False:
        print("Error",req.status_code, ":", req.reason)
        return None
    GTFSLast = req.headers['Last-Modified'] # Tue, 23 Apr 2019 15:18:38 GMT
    GTFSdatetime = datetime.datetime.strptime(GTFSLast, '%a, %d %b %Y %H:%M:%S GMT')
    return GTFSdatetime

def downloadFile(url, filename):
    #Downloads the TEC-GTFS file
    GTFSfile = requests.get(url, stream=True)
    with open(filename, "wb") as f:
        f.write(GTFSfile.content)

def unzipFile(filename, folder):
    #Unzips the file
    with zipfile.ZipFile(filename, "r") as GTFSzip:
        GTFSzip.extractall(folder)

def loadCSV(filename, folder):
    os.system("""mysql -h%s -u%s -p%s "%s" -e "LOAD DATA LOCAL INFILE '%s/%s.txt' INTO TABLE %s FIELDS TERMINATED BY ',' ENCLOSED BY '\\"' LINES TERMINATED BY '\r\n' IGNORE 1 LINES; COMMIT;" """ % (config['DB_HOSTNAME'],config['DB_USER'], config['DB_PASSWORD'],config['DB_TMP'],folder,filename, filename))

def emptyTable(tableName):
    curs.execute("DELETE FROM " + tableName)

def migrateTable(tableName):
    curs.execute("DELETE FROM " + config['DB_MAIN'] + "." + tableName)
    curs.execute("INSERT INTO " + config['DB_MAIN'] + "." + tableName +
                 " SELECT * FROM " + tableName)


if __name__ == '__main__':
    lastGTFS = getLastGTFS(config['URL'])
    lastUpdate = getLastUpdate()
    print("Last gtfs timestamp :",lastGTFS)
    print("Last update timestamp :", lastUpdate)
    if lastUpdate is not None and lastGTFS < lastUpdate:
        print("Database already up to date")
        sys.exit(0)
    print("Will be updating database")
    os.chdir("/tmp")
    print("Downloading GTFS Data")
    downloadFile(config['URL'], config['ZIP_NAME'])
    print("Decompressing GTFS data")
    unzipFile(config['ZIP_NAME'], config['UNZIP_FOLDER'])
    files = ["calendar", "calendar_dates", "shapes", "routes", "trips", "stop_times"]
    for filename in files:
        print(f"Emptying {filename} first")
        emptyTable(filename)
        db.commit()
        print(f"Importing {filename}")
        loadCSV(filename, config['UNZIP_FOLDER'])
        db.commit()
    print("Emptying trip_times")
    emptyTable("trip_times")
    print("Generating trip_times")
    curs.execute("INSERT INTO trip_times SELECT trip_id, min(arrival_time), max(departure_time) FROM stop_times GROUP BY trip_id")
    db.commit()
    files.append("trip_times")
    for filename in files:
        print(f"Migrating {filename}")
        migrateTable(filename)
    curs.execute("INSERT INTO " + config['DB_MAIN'] + ".meta VALUES (1,%s) ON DUPLICATE KEY UPDATE lastUpdate = %s", (datetime.datetime.now(),datetime.datetime.now()))
    db.commit()
    os.unlink(config['ZIP_NAME'])
    shutil.rmtree(config['UNZIP_FOLDER'])

