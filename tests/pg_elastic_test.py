#!/usr/bin/env python3

import os
import logging
import elasticsearch.exceptions
from datetime import datetime
from elasticsearch_dsl import DocType, Date, Integer, Keyword, Text, Search
from elasticsearch_dsl.connections import connections

PORT = int(os.getenv('ELASTIC_PORT', 9200))

class Tweet(DocType):
    message = Text(analyzer='english')
    user = Text()
    post_date = Date()

    class Meta:
        index = 'twitter'
        doc_type = 'tweet'

class TestBasic:
    def setup_class(self):
        connections.create_connection(hosts=['localhost'], port=PORT)
        Tweet.init()

    def test_index(self):
        message = 'trying out Elasticsearch'
        user = 'kimchy'
        tweet = Tweet(meta={'id': 1}, message=message, user=user, post_date = datetime.now())
        tweet.save()
        try:
            tweet = Tweet.get(id = 1)
            assert(tweet.message == message)
            assert(tweet.user == user)
        except NotFoundError:
            assert(false)

    def test_search(self):
        s = Search(index="twitter") \
            .query("match", message="try")
        response = s.execute()
        assert(response.hits.total == 1)

        for hit in response:
            assert(hit.message == 'trying out Elasticsearch')

        s = Search(index="twitter") \
            .query("match", message="trying")
        response = s.execute()
        assert(response.hits.total == 1)

        s = Search(index="twitter") \
            .query("match", message="elastic")
        response = s.execute()
        assert(response.hits.total == 0)

    def test_health(self):
        health = connections.get_connection().cluster.health()
        assert(health['status'] == 'yellow' or health['status'] == 'green')

