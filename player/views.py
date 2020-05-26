from django.shortcuts import render, redirect
from django.http import HttpResponse
from django.contrib.auth import logout

def index(request):
    if request.user.is_authenticated:
        return render(request, 'player/index.html')
    else:
        return render(request, 'player/login.html')

def logout_view(request):
    logout(request)
    return redirect(index)
