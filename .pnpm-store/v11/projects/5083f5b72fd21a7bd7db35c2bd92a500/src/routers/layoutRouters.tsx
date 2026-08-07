import { Navigate } from 'react-router-dom'
import type { RouteObject } from 'react-router-dom'
import { Layout } from '@/components/core/layout'
import type { AppRouteObject } from '@/types/router'
import AuthGuard from './AuthGuard'
import RouteFallback from './RouteFallback'
import { staticRoutes } from './staticRoutes'

const layoutStaticRoutes: RouteObject[] = staticRoutes
  .filter((route) => route.isLayout)
  .map((route) => ({
    path: route.path,
    element: route.element,
    handle: {
      name: route.name,
      meta: route.meta,
    },
  }))

export const fullScreenRoutes: RouteObject[] = staticRoutes
  .filter((route) => !route.isLayout)
  .map((route) => ({
    path: route.path,
    element: route.element,
    handle: {
      name: route.name,
      meta: route.meta,
    },
  }))

const isLayoutRoute = (route: AppRouteObject) =>
  route.handle?.meta?.layout !== false

export const getFullScreenDynamicRoutes = (dynamicRoutes: AppRouteObject[]): RouteObject[] =>
  dynamicRoutes
    .filter((route) => !isLayoutRoute(route))
    .map((route) => ({
      ...route,
      element: <AuthGuard>{route.element}</AuthGuard>,
    }))

export const getLayoutRouters = (dynamicRoutes: AppRouteObject[]): RouteObject[] => [
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      ...layoutStaticRoutes,
      ...dynamicRoutes.filter(isLayoutRoute),
      { path: '*', element: <RouteFallback /> },
    ],
  },
]
