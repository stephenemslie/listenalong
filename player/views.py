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
    room = Room.objects.create(user=request.user)
    redirect('room-detail', slug=room.slug)


class RoomDetailView(DetailView):

    model = Room
