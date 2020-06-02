import time

from django.core.management.base import BaseCommand
from django.conf import settings
import schedule

from player.models import Room


class Command(BaseCommand):

    def update_rooms(self):
        # select rooms with an owner
        rooms = Room.objects.filter(user__room_owner=True)
        print(f'Updating {rooms.count()} rooms')
        for room in rooms:
            resync_needed = room.update_progress()
            room.drop_inactive_members()
            if resync_needed:
                room.sync_members()

    def handle(self, *args, **options):
        count_seconds = settings.SPOTIFY_POLLING_INTERVAL_SECONDS
        schedule.every(count_seconds).seconds.do(self.update_rooms)
        print(f'Updating rooms every {count_seconds} seconds')
        self.update_rooms()
        while True:
            schedule.run_pending()
            time.sleep(1)
