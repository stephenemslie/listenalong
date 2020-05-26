from django.urls import path

from . import views


urlpatterns = [
    path('', views.index),
    path('logout/', views.logout_view, name="player-views-logout"),
]
