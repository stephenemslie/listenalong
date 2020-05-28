import time

from django.core.management.base import BaseCommand
import schedule

from player.models import Room


class Command(BaseCommand):

    def update_rooms(self):
        # select rooms with an owner
        rooms = Room.objects.filter(user__room_owner=True)
        print(f'Updating {rooms.count()} rooms')
        for room in rooms:
            room.update_progress()
            room.drop_inactive_members()

    def handle(self, *args, **options):
        count_seconds = 15
        schedule.every(count_seconds).seconds.do(self.update_rooms)
        print(f'Updating rooms every {count_seconds} seconds')
        while True:
            schedule.run_pending()
            time.sleep(1)
