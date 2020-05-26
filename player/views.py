from django.shortcuts import render
from django.http import HttpResponse

def index(request):
    if request.user.is_authenticated:
        return render(request, 'player/index.html')
    else:
        return render(request, 'player/login.html')
