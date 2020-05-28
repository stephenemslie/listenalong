import time
import string

from random import SystemRandom

from django.db import models
from django.contrib.auth.models import AbstractUser
from social_django.utils import load_strategy
from django.utils import timezone
import tekore as tk


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

    def spotify_is_active(self, threshold=5):
        """Return True if the user should be considered active.

        An active user:
         - has an active device
         - is currently playing the room's context
         - has progress_ms within threshold seconds of  room's progress_ms
        """
        spotify = tk.Spotify(self.get_spotify_token())
        playing = spotify.playback_currently_playing()
        room = self.room
        if playing is None:
            return False
        if playing.is_playing is False:
            return False
        if playing.context is None:
            return False
        if playing.context.uri != room.context_uri:
            return False
        if playing.item.uri != room.item_uri:
            return False
        if abs(playing.progress_ms - room.progress_ms) > threshold:
            return False
        return True

    def spotify_sync(self):
        """Sync this user's player with their room"""
        room = self.room
        progress_ms = room.adjust_progress()
        spotify = tk.Spotify(self.get_spotify_token())
        spotify.playback_start_context(
            room.context_uri,
            room.item_id,
            progress_ms
        )


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
        return f'{self.slug}'

    @property
    def owner(self):
        return self.user_set.get(room_owner=True)

    def adjust_progress(self):
        """Return room's adjusted progress, in milliseconds.

        progress_ms is adjusted for the time elapsed since last sync.
        """
        # TODO: if adjusted_progress_ms > track length, adjust into next track
        elapsed = time.time() * 1000 - self.timestamp.timestamp() * 1000
        return self.progress_ms + elapsed + 1000

    def update_progress(self):
        spotify = tk.Spotify(self.owner.get_spotify_token())
        self.timestamp = timezone.now()
        playing = spotify.playback_currently_playing()
        if playing is None or playing.context is None:
            return

        self.is_playing = playing.is_playing
        self.progress_ms = playing.progress_ms

        self.item_id = playing.item.id
        self.item_uri = playing.item.uri
        self.item_name = playing.item.name

        self.context_uri = playing.context.uri
        self.context_type = playing.context.type.value
        self.save()
