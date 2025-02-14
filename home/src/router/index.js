import { createRouter, createWebHashHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";

const routes = [
  {
    path: "/",
    name: "home",
    component: HomeView,
  },
  {
    path: "/about",
    name: "about",
    component: () => import("../views/AboutView.vue"),
  },
  {
    path: "/departamentos1",
    name: "departamentos1",
    component: () => import("../views/Departamento1View.vue"),
  },
  {
    path: "/departamentos2",
    name: "departamentos2",
    component: () => import("../views/Departamento2View.vue"),
  },
  {
    path: "/departamentos3",
    name: "departamentos3",
    component: () => import("../views/Departamento3View.vue"),
  },
  {
    path: "/departamentos4",
    name: "departamentos4",
    component: () => import("../views/Departamento4View.vue"),
  },
  {
    path: "/noticias",
    name: "noticias",
    component: () => import("../views/NoticiasView.vue"),
  },
  {
    path: "/programas1",
    name: "programas1",
    component: () => import("../views/Programas1View.vue"),
  },
  {
    path: "/programas2",
    name: "programas2",
    component: () => import("../views/Programas2View.vue"),
  },
  {
    path: "/programas3",
    name: "programas3",
    component: () => import("../views/Programas3View.vue"),
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
