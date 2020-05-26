from django.urls import path

from . import views


urlpatterns = [
    path('', views.index),
    path('rooms/', views.room_create_view, name='room-create'),
    path('rooms/<slug:slug>/',
         views.RoomDetailView.as_view(),
         name='room-detail'),
    path('logout/', views.logout_view, name='player-views-logout'),
]
