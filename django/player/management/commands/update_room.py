from django.core.management.base import BaseCommand, CommandError
from player.models import Room


class Command(BaseCommand):

    def add_arguments(self, parser):
        parser.add_argument('--id', type=int)
        parser.add_argument('--all', action='store_true',
                            help='update all active rooms')
        parser.add_argument('--sync-members', action='store_true',
                            help='sync room members')

    def handle(self, *args, **options):
        if options['all'] and options['id']:
            raise CommandError('Specify only one of --all or --id')
        if options['all']:
            rooms = Room.objects.filter(user__room_owner=True)
        if options['id']:
            rooms = [Room.objects.get(id=options['id'])]
        for room in rooms:
            room.update_progress()
            if options['sync_members']:
                room.sync_members()
