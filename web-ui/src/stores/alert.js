import { defineStore } from 'pinia'
import { ref } from 'vue'
import { io } from 'socket.io-client'

export const useAlertStore = defineStore('alert', () => {
  const socket = ref(null)
  const alerts = ref([])
  const connected = ref(false)

  const connectWebSocket = () => {
    socket.value = io('http://localhost:5000', {
      transports: ['websocket']
    })

    socket.value.on('connect', () => {
      connected.value = true
      console.log('WebSocket connected')
    })

    socket.value.on('disconnect', () => {
      connected.value = false
      console.log('WebSocket disconnected')
    })

    socket.value.on('new_alert', (alert) => {
      alerts.value.unshift(alert)
      if (alerts.value.length > 100) {
        alerts.value.pop()
      }
    })
  }

  const disconnect = () => {
    if (socket.value) {
      socket.value.disconnect()
    }
  }

  return {
    socket,
    alerts,
    connected,
    connectWebSocket,
    disconnect
  }
})
