from django.shortcuts import render, redirect
from django.http import HttpResponse
from django.contrib.auth import logout
from django.contrib.auth.decorators import login_required
from django.views.generic import CreateView
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


@method_decorator(login_required, name='dispatch')
class RoomCreateView(CreateView):

    model = Room

