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

class Room(models.Model):
    slug = models.SlugField(default=generate_room_slug)

    def __str__(self):
        return f'{self.slug} ({self.user.username})'

