CREATE TABLE IF NOT EXISTS omni_dash.illuminance(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash.humidity(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash.temperature(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash.moved(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash.sbothumidity(
   Time DATETIME(0),
   Value INT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash.sbottemperature(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash.co2ppm(
   Time DATETIME(0),
   Value INT,
   PRIMARY KEY (Time)
);