CREATE TABLE IF NOT EXISTS omni_dash_dev.illuminance(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash_dev.humidity(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash_dev.temperature(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash_dev.moved(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash_dev.sbothumidity(
   Time DATETIME(0),
   Value INT,
   PRIMARY KEY (Time)
);

CREATE TABLE IF NOT EXISTS omni_dash_dev.sbottemperature(
   Time DATETIME(0),
   Value FLOAT,
   PRIMARY KEY (Time)
);