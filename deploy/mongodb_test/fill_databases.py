#!/usr/bin/python2
from pymongo import MongoClient
from multiprocessing import Pool


connection = MongoClient('localhost', maxPoolSize=200)
	

def insert(numbase):
	db_name = 'backup_me_{}'.format(numbase)
	db = connection[db_name]
	bulk = list()
	for num_coll in xrange(0,5):
		coll_name = 'user_data_{}'.format(num_coll)
		db[coll_name].insert({"1":"1"})



if __name__ == '__main__':
	p = Pool(3)
	p.map(insert, [x for x in xrange(0,20)])
	p.close()
	p.join()

