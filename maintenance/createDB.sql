USE TEC;

DROP TABLE IF EXISTS agency;

CREATE TABLE `agency` (
  agency_id CHAR(1) PRIMARY KEY,
  agency_name VARCHAR(255),
  agency_url VARCHAR(255),
  agency_timezone VARCHAR(50),
  agency_lang CHAR(2),
  agency_phone VARCHAR(50)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS calendar;

CREATE TABLE `calendar` (
  service_id VARCHAR(64),
  monday TINYINT(1),
  tuesday TINYINT(1),
  wednesday TINYINT(1),
  thursday TINYINT(1),
  friday TINYINT(1),
  saturday TINYINT(1),
  sunday TINYINT(1),
  start_date DATE,
  end_date DATE,
  KEY `service_id` (service_id)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS calendar_dates;

CREATE TABLE `calendar_dates` (
  service_id VARCHAR(64),
  `date` DATE,
  exception_type INT(2),
  KEY `service_id` (service_id),
  KEY `exception_type` (exception_type)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS routes;

CREATE TABLE `routes` (
  route_id CHAR(11) PRIMARY KEY,
  agency_id CHAR(1),
  route_short_name VARCHAR(50),
  route_long_name VARCHAR(255),
  route_desc VARCHAR(255),
  route_type INT(2),
  route_url VARCHAR(255),
  KEY `route_type` (route_type)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS shapes;

CREATE TABLE `shapes` (
  shape_id VARCHAR(50),
  shape_pt_lat DECIMAL(9,6),
  shape_pt_lon DECIMAL(9,6),
  shape_pt_sequence INT(11),
  KEY `shape_id` (shape_id)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS stop_times;

CREATE TABLE `stop_times` (
  trip_id VARCHAR(64),
  arrival_time TIME,
  departure_time TIME,
  stop_id CHAR(8),
  stop_sequence INT(11),
  pickup_type INT(11),
  drop_off_type INT(11),
  KEY `trip_id` (trip_id),
  KEY `stop_id` (stop_id),
  KEY `stop_sequence` (stop_sequence)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS stops;

CREATE TABLE `stops` (
  stop_id CHAR(8) PRIMARY KEY,
  stop_code VARCHAR(64),
  stop_name VARCHAR(255),
  stop_desc VARCHAR(255),
  stop_lat DECIMAL(9,6),
  stop_lon DECIMAL(9,6),
  zone_id INT(11),
  stop_url VARCHAR(255),
  location_type INT(11),
  KEY `stop_lat` (stop_lat),
  KEY `stop_lon` (stop_lon)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS trips;

CREATE TABLE `trips` (
  route_id CHAR(11),
  service_id VARCHAR(64),
  trip_id VARCHAR(64) PRIMARY KEY,
  trip_short_name VARCHAR(255),
  direction_id TINYINT(1),
  block_id INT(11),
  shape_id VARCHAR(50),
  KEY `route_id` (route_id),
  KEY `service_id` (service_id),
  KEY `direction_id` (direction_id),
  KEY `shape_id` (shape_id)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS trip_times;
CREATE TABLE `trip_times` (
  trip_id VARCHAR(64),
  begin_time TIME,
  end_time TIME,
  KEY `trip_id` (trip_id),
  KEY `begin_time` (begin_time),
  KEY `end_time` (end_time)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS meta;
CREATE TABLE `meta` (
  id INT(11) PRIMARY KEY,
  lastUpdate DATETIME
)
COLLATE 'utf8mb4_bin';


USE TEC_TMP;

DROP TABLE IF EXISTS agency;

CREATE TABLE `agency` (
  agency_id CHAR(1) PRIMARY KEY,
  agency_name VARCHAR(255),
  agency_url VARCHAR(255),
  agency_timezone VARCHAR(50),
  agency_lang CHAR(2),
  agency_phone VARCHAR(50)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS calendar;

CREATE TABLE `calendar` (
  service_id VARCHAR(64),
  monday TINYINT(1),
  tuesday TINYINT(1),
  wednesday TINYINT(1),
  thursday TINYINT(1),
  friday TINYINT(1),
  saturday TINYINT(1),
  sunday TINYINT(1),
  start_date DATE,
  end_date DATE,
  KEY `service_id` (service_id)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS calendar_dates;

CREATE TABLE `calendar_dates` (
  service_id VARCHAR(64),
  `date` DATE,
  exception_type INT(2),
  KEY `service_id` (service_id),
  KEY `exception_type` (exception_type)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS routes;

CREATE TABLE `routes` (
  route_id CHAR(11) PRIMARY KEY,
  agency_id CHAR(1),
  route_short_name VARCHAR(50),
  route_long_name VARCHAR(255),
  route_desc VARCHAR(255),
  route_type INT(2),
  route_url VARCHAR(255),
  KEY `route_type` (route_type)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS shapes;

CREATE TABLE `shapes` (
  shape_id VARCHAR(50),
  shape_pt_lat DECIMAL(9,6),
  shape_pt_lon DECIMAL(9,6),
  shape_pt_sequence INT(11),
  KEY `shape_id` (shape_id)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS stop_times;

CREATE TABLE `stop_times` (
  trip_id VARCHAR(64),
  arrival_time TIME,
  departure_time TIME,
  stop_id CHAR(8),
  stop_sequence INT(11),
  pickup_type INT(11),
  drop_off_type INT(11),
  KEY `trip_id` (trip_id),
  KEY `stop_id` (stop_id),
  KEY `stop_sequence` (stop_sequence)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS stops;

CREATE TABLE `stops` (
  stop_id CHAR(8) PRIMARY KEY,
  stop_code VARCHAR(64),
  stop_name VARCHAR(255),
  stop_desc VARCHAR(255),
  stop_lat DECIMAL(9,6),
  stop_lon DECIMAL(9,6),
  zone_id INT(11),
  stop_url VARCHAR(255),
  location_type INT(11),
  KEY `stop_lat` (stop_lat),
  KEY `stop_lon` (stop_lon)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS trips;

CREATE TABLE `trips` (
  route_id CHAR(11),
  service_id VARCHAR(64),
  trip_id VARCHAR(64) PRIMARY KEY,
  trip_short_name VARCHAR(255),
  direction_id TINYINT(1),
  block_id INT(11),
  shape_id VARCHAR(50),
  KEY `route_id` (route_id),
  KEY `service_id` (service_id),
  KEY `direction_id` (direction_id),
  KEY `shape_id` (shape_id)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS trip_times;
CREATE TABLE `trip_times` (
  trip_id VARCHAR(64),
  begin_time TIME,
  end_time TIME,
  KEY `trip_id` (trip_id),
  KEY `begin_time` (begin_time),
  KEY `end_time` (end_time)
)
COLLATE 'utf8mb4_bin';

DROP TABLE IF EXISTS meta;
CREATE TABLE `meta` (
  id INT(11) PRIMARY KEY,
  lastUpdate DATETIME
)
COLLATE 'utf8mb4_bin';
