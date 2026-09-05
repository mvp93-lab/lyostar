import { ref } from 'vue'

const toast = ref({
  visible: false,
  message: '',
  type: 'success', // 'success' | 'info' | 'error'
  timeoutId: null
})

export function useToast() {
  function showToast(message, type = 'success', duration = 3000) {
    if (toast.value.timeoutId) {
      clearTimeout(toast.value.timeoutId)
    }
    toast.value = {
      visible: true,
      message,
      type,
      timeoutId: setTimeout(() => {
        toast.value.visible = false
      }, duration)
    }
  }

  function hideToast() {
    if (toast.value.timeoutId) {
      clearTimeout(toast.value.timeoutId)
    }
    toast.value.visible = false
  }

  return {
    toast,
    showToast,
    hideToast
  }
}
