from bs4 import BeautifulSoup
import requests
from flask import Flask, abort, jsonify
import json

ZONE01_URL = "https://www.zone01.be/hercules/resultaat_uitgebreid"
#bus_id = 5736


app = Flask(__name__)

@app.errorhandler(404)
def resource_not_found(e):
    return jsonify(error=str(e)), 404

@app.route('/find/<bus_id>')
def bus(bus_id):
    data = {
        "owner": "OTW",
        "busnr": bus_id,
        "status":"A",
        "submit": "Zoeken"
    }
    resp = requests.post(url=ZONE01_URL, data=data)
    soup = BeautifulSoup(resp.content, features="lxml")
    bus_table = soup.find(id="herculesdetails")
    if bus_table is None:
        abort(404, "Bus ID not found")
    busses = bus_table.findAll("tr")
    for bus in busses:
       bus_owner = bus.find("span", class_="owners")
       bus_nr = bus.find("span", class_="busnr")
       license_plate = bus.find("span", class_="kenteken")
       href = bus.find("a").get("href")
       if "OTW" in bus_owner.text and bus_nr.text == str(bus_id):
           return jsonify(license_plate=license_plate.text, link="https://www.zone01.be/hercules/"+str(href))
    abort(404, "Bus ID not found")
