CREATE TABLE pricelist (distributorId INT NOT NULL, art INT NOT NULL, count INT, price INT, CONSTRAINT id PRIMARY KEY (distributorId, art))