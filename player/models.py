import string
from random import SystemRandom

from django.db import models
from django.contrib.auth.models import AbstractUser


def generate_room_slug():
    choice = SystemRandom().choice
    chars = string.ascii_uppercase + string.digits
    return "".join(choice(chars) for i in range(6))


class User(AbstractUser):

    room = models.ForeignKey(
        'Room', on_delete=models.PROTECT, blank=True, null=True)
    room_owner = models.BooleanField(default=False)

    def get_spotify_token(self):
        social_user = self.social_auth.get(provider='spotify')
        expires = social_user.extra_data['auth_time'] + \
            social_user.extra_data['expires_in']
        if expires <= int(time.time()):
            social_user.refresh_token(load_strategy())
            social_user.save()
        return social_user.access_token

class Room(models.Model):
    slug = models.SlugField(default=generate_room_slug)

    timestamp = models.DateTimeField(blank=True, null=True)

    is_playing = models.BooleanField(default=False)
    progress_ms = models.FloatField(blank=True, null=True)

    context_uri = models.TextField(blank=True, null=True)
    context_type = models.TextField(blank=True, null=True)

    item_id = models.TextField(blank=True, null=True)
    item_uri = models.TextField(blank=True, null=True)
    item_name = models.TextField(blank=True, null=True)

    def __str__(self):
        return f'{self.slug} ({self.user.username})'

