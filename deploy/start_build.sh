#!/bin/bash
apt-get install -y python-pip && pip install pymongo
docker-compose -f shkaff.yml down && docker-compose -f shkaff.yml up -d && ./mongodb_test/fill_databases.py
