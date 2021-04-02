from bs4 import BeautifulSoup
import requests
from flask import Flask, abort, jsonify
import json
import re

ZONE01_URL = "https://www.zone01.be/zoeken"
#bus_id = 5736


app = Flask(__name__)

@app.errorhandler(404)
def resource_not_found(e):
    return jsonify(error=str(e)), 404

@app.route('/find/<bus_id>')
def bus(bus_id):
    resp = requests.get(url=ZONE01_URL, params={"s":bus_id})
    soup = BeautifulSoup(resp.content, features="lxml")
    bus_table = soup.find("table",id="example")
    if bus_table is None:
        abort(404, "Bus ID not found")
    busses = bus_table.find_all(id="herculesdetails")
    for bus in busses:
        details = bus.find_all("tr")
        for detail in details:
            if not detail.find(class_="owners",string=re.compile("OTW\s\(.+\)")):
                continue
            if not detail.find(class_="busnr", string=str(bus_id)):
                continue
            if not detail.find(alt="status",src=re.compile("/A.png$")):
                continue
            license_plate = detail.find("span", class_="kenteken")
            href = detail.find("a").get("href")
            return jsonify(license_plate=license_plate.text, link="https://www.zone01.be"+str(href))
    abort(404, "Bus ID not found")
