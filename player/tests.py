import time
from unittest.mock import patch, MagicMock

from django.db.models.signals import pre_save, post_save
from django.test import TestCase
from django.utils import timezone
from factory.django import DjangoModelFactory
import factory
import tekore

from player.models import Room


class SocialAuthFactory(factory.django.DjangoModelFactory):
    class Meta:
        model = 'social_django.UserSocialAuth'

    user = factory.SubFactory('player.tests.UserFactory', social_auth=None)
    provider = 'spotify'
    extra_data = {
        'auth_time': time.time(),
        'expires_in': 10000
    }


class UserFactory(factory.django.DjangoModelFactory):
    class Meta:
        model = 'player.User'

    social_auth = factory.RelatedFactory(SocialAuthFactory, 'user')


class RoomFactory(factory.django.DjangoModelFactory):
    class Meta:
        model = 'player.Room'

    progress_ms = 5000
    timestamp = factory.LazyAttribute(lambda a: timezone.now())


class SpotifyContextTypeFactory(factory.Factory):

    class Meta:
        model = MagicMock

    value = "playlist"


class SpotifyContextFactory(factory.Factory):

    class Meta:
        model = MagicMock

    uri = factory.Faker('uuid4')
    type = factory.SubFactory(SpotifyContextTypeFactory)


class SpotifyItemFactory(factory.Factory):

    class Meta:
        model = MagicMock

    id = factory.Faker('uuid4')
    uri = factory.Faker('uuid4')
    _name = factory.Faker('pystr')

    # MagicMock reserves `name` for the constructor, so we need to set it post
    # generation
    # https://docs.python.org/3/library/unittest.mock.html#mock-names-and-the-name-attribute
    @factory.post_generation
    def set_name(obj, create, extracted, **kwargs):
        obj.name = obj._name


class SpotifyPlayingFactory(factory.Factory):

    class Meta:
        model = MagicMock

    is_playing = False
    progress_ms = 1000
    item = factory.SubFactory(SpotifyItemFactory)
    context = factory.SubFactory(SpotifyContextFactory)


class UserTestCase(TestCase):

    @patch.object(tekore.Spotify, 'playback_shuffle')
    @patch.object(tekore.Spotify, 'playback_devices')
    def test_disable_shuffle_on_own(self, playback_devices, playback_shuffle):
        """Test that owner's shuffle is disabled when they create a room."""
        playback_devices.return_value = [MagicMock()]
        user = UserFactory()
        room = RoomFactory()
        user.room = room
        user.room_owner = True
        user.save()
        device_id = playback_devices()[0].id
        playback_shuffle.assert_called_once_with(False, device_id)
