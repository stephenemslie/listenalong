import os
import json

import boto3
from botocore.exceptions import ClientError
import environ


env = environ.Env(
    DEBUG=(bool, False),
    ALLOWED_HOSTS=(list, '127.0.0.1'),
    SPOTIFY_POLLING_INTERVAL_SECONDS=(int, 15),
    USE_X_FORWARDED_HOST=(bool, False),
    SECURE_SSL_REDIRECT=(bool, False),
    SOCIAL_AUTH_SPOTIFY_KEY=(
        str, "secret://listenalong-django/SOCIAL_AUTH_SPOTIFY_KEY"),
    SOCIAL_AUTH_SPOTIFY_SECRET=(
        str, "secret://listenalong-django/SOCIAL_AUTH_SPOTIFY_SECRET"),
    SECRET_KEY=(str, "secret://listenalong-django/SOCIAL_AUTH_SPOTIFY_SECRET")
)

environ.Env.read_env('/usr/src/app/.env')


def resolve_secrets(env):
    session = boto3.session.Session()
    client = session.client(
        service_name='secretsmanager',
        region_name='eu-west-1'
    )
    for key in env.scheme.keys():
        value = env(key)
        if hasattr(value, 'startswith') and value.startswith('secret://'):
            name = value.split('secret://', 1)[1]
            secret_name, jsonkey = name.split('/', 1)
            try:
                response = client.get_secret_value(SecretId=secret_name)
                json_value = response['SecretString']
            except ClientError as e:
                if e.response['Error']['Code'] == 'ResourceNotFoundException':
                    print(f"The requested secret {name} was not found")
                raise
            secret_value = json.loads(json_value)
            env.ENVIRON[key] = secret_value[jsonkey]


resolve_secrets(env)

# Build paths inside the project like this: os.path.join(BASE_DIR, ...)
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


# Quick-start development settings - unsuitable for production
# See https://docs.djangoproject.com/en/3.0/howto/deployment/checklist/

SECRET_KEY = env('SECRET_KEY')

# SECURITY WARNING: don't run with debug turned on in production!
DEBUG = env('DEBUG')

ALLOWED_HOSTS = env('ALLOWED_HOSTS')
USE_X_FORWARDED_HOST = env('USE_X_FORWARDED_HOST')
SECURE_SSL_REDIRECT = env('SECURE_SSL_REDIRECT')
SECURE_PROXY_SSL_HEADER = ('HTTP_X_FORWARDED_PROTO', 'https')


# Application definition

INSTALLED_APPS = [
    'django.contrib.admin',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.messages',
    'django.contrib.staticfiles',
    'django_s3_sqlite',
    'social_django',
    'player',
]

MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.common.CommonMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'django.contrib.messages.middleware.MessageMiddleware',
    'django.middleware.clickjacking.XFrameOptionsMiddleware',
]

ROOT_URLCONF = 'listenalong.urls'

TEMPLATES = [
    {
        'BACKEND': 'django.template.backends.django.DjangoTemplates',
        'DIRS': [],
        'APP_DIRS': True,
        'OPTIONS': {
            'context_processors': [
                'django.template.context_processors.debug',
                'django.template.context_processors.request',
                'django.contrib.auth.context_processors.auth',
                'django.contrib.messages.context_processors.messages',
                'social_django.context_processors.backends',
                'social_django.context_processors.login_redirect',
            ],
        },
    },
]

WSGI_APPLICATION = 'listenalong.wsgi.application'


# Database
# https://docs.djangoproject.com/en/3.0/ref/settings/#databases

DATABASES = {
    'default': env.db()
}


# Password validation
# https://docs.djangoproject.com/en/3.0/ref/settings/#auth-password-validators

AUTH_PASSWORD_VALIDATORS = [
    {
        'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator',
    },
]


# Internationalization
# https://docs.djangoproject.com/en/3.0/topics/i18n/

LANGUAGE_CODE = 'en-us'

TIME_ZONE = 'UTC'

USE_I18N = True

USE_L10N = True

USE_TZ = True


# Static files (CSS, JavaScript, Images)
# https://docs.djangoproject.com/en/3.0/howto/static-files/

STATIC_URL = '/static/'

LOGIN_URL = '/'
LOGIN_REDIRECT_URL = '/'

AUTH_USER_MODEL = 'player.User'

# Social Auth
SOCIAL_AUTH_URL_NAMESPACE = 'social'

AUTHENTICATION_BACKENDS = [
    'django.contrib.auth.backends.ModelBackend',
    'social_core.backends.spotify.SpotifyOAuth2',
]

SOCIAL_AUTH_SPOTIFY_KEY = env('SOCIAL_AUTH_SPOTIFY_KEY')
SOCIAL_AUTH_SPOTIFY_SECRET = env('SOCIAL_AUTH_SPOTIFY_SECRET')
SOCIAL_AUTH_SPOTIFY_SCOPE = [
    'user-read-playback-state',
    'user-modify-playback-state',
    'user-read-currently-playing',
    'playlist-read-collaborative'
]
SOCIAL_AUTH_SPOTIFY_GET_ALL_EXTRA_DATA = True

# Player app specific

SPOTIFY_POLLING_INTERVAL_SECONDS = env('SPOTIFY_POLLING_INTERVAL_SECONDS', 15)
