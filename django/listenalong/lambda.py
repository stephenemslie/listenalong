import serverless_wsgi
from .wsgi import application as app


def handler(event, context):
    return serverless_wsgi.handle_request(app.app, event, context)
