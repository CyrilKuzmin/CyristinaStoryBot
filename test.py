from pymongo import *

client = MongoClient('mongodb://127.0.0.1:27017')
db = client['stories']
col = db['stories']
for i in col.find({}, {"Title": 1}):
    print(i['Title'])