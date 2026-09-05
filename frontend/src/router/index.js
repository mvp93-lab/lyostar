import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const routes = [
  {
    path: '/',
    redirect: '/books'
  },
  {
    path: '/books',
    name: 'books'
  },
  {
    path: '/continue-reading',
    name: 'continue-reading'
  },
  {
    path: '/tags',
    name: 'tags'
  },
  {
    path: '/tags/:tag',
    name: 'tag-filter'
  },
  {
    path: '/authors',
    name: 'authors'
  },
  {
    path: '/authors/:author',
    name: 'author-books'
  },
  {
    path: '/series',
    name: 'series'
  },
  {
    path: '/series/:series',
    name: 'series-books'
  },
  {
    path: '/shelves/:id',
    name: 'shelf-books'
  },
  {
    path: '/read/:id',
    name: 'reader'
  },
  {
    path: '/books/:id',
    name: 'book-detail'
  },
  {
    path: '/search',
    name: 'search'
  },
  {
    path: '/users',
    name: 'users',
    meta: { requiresAdmin: true }
  },
  {
    path: '/upload',
    name: 'upload'
  },
  {
    path: '/login',
    name: 'login'
  },
  {
    path: '/setup',
    name: 'setup'
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/books'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from) => {
  const { isAuthenticated, isAdmin, setupRequired, loading, checkAuth } = useAuth()

  if (loading.value) {
    await checkAuth()
  }

  if (setupRequired.value) {
    if (to.name !== 'setup') {
      return { name: 'setup' }
    }
    return true
  }

  if (to.name === 'setup') {
    return { name: 'books' }
  }

  if (!isAuthenticated.value) {
    if (to.name !== 'login') {
      return {
        name: 'login',
        query: to.fullPath !== '/books' ? { redirect: to.fullPath } : undefined
      }
    }
    return true
  }

  if (to.name === 'login') {
    const redirect = to.query.redirect
    if (typeof redirect === 'string' && redirect.startsWith('/')) {
      return redirect
    }
    return { name: 'books' }
  }

  if (to.meta.requiresAdmin && !isAdmin.value) {
    return { name: 'books' }
  }

  return true
})

export default router
