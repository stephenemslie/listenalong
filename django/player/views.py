from django.shortcuts import render, redirect
from django.http import HttpResponse
from django.contrib.auth import logout
from django.contrib.auth.decorators import login_required
from django.views.generic import DetailView
from django.views.decorators.http import require_http_methods
from django.utils.decorators import method_decorator

from .models import Room


def index(request):
    if request.user.is_authenticated:
        return render(request, 'player/index.html')
    else:
        return render(request, 'player/login.html')


def logout_view(request):
    logout(request)
    return redirect(index)


@login_required
def room_create_view(request):
    user = request.user
    room = Room.objects.create(user=request.user)
    user.room = room
    user.room_owner = True
    user.save()
    return redirect('room-detail', slug=room.slug)


@login_required
def room_join_view(request, slug):
    room = Room.objects.get(slug=slug)
    user = request.user
    user.room = room
    user.room_owner = False
    user.save()
    user.spotify_sync()
    return redirect('room-detail', slug=room.slug)


@login_required
def room_leave_view(request):
    user = request.user
    user.room = None
    user.room_owner = False
    user.save()
    return redirect('index')


@method_decorator(login_required, name='dispatch')
class RoomDetailView(DetailView):

    model = Room
