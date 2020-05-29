import { BreadCrumbs } from "/_dist_/BreadCrumbs.js";
import { RootView } from "/_dist_/root/RootView.js";
import { GenreView } from "/_dist_/genre/GenreView.js";
import { GenreTableView } from "/_dist_/genre/GenreTableView.js";
// import { GenreDetailView } from "/_dist_/genre/GenreDetailView.js";
import { SerieView } from "/_dist_/serie/SerieView.js";
import { SerieTableView } from "/_dist_/serie/SerieTableView.js";
// import { SerieDetailView } from "/_dist_/serie/SerieDetailView.js";
import { MovieView } from "/_dist_/movie/MovieView.js";
import { MovieTableView } from "/_dist_/movie/MovieTableView.js";
// import { MovieDetailView } from "/_dist_/movie/MovieDetailView.js";
import { LanguageTableView } from "/_dist_/language/LanguageTableView.js";
// import { LanguageDetailView } from "/_dist_/language/LanguageDetailView.js";
import { SearchView } from "/_dist_/search/SearchView.js";
import { UnknownView } from "/_dist_/unknown/UnknownView.js";
import { AboutView } from "/_dist_/about/AboutView.js";

import { Cache } from "/_dist_/Cache.js";
import { Router } from '/_dist_/lib/@vaadin/router.js';
import "/_dist_/search/SearchFieldService.js";

new Cache().clear();

const outlet = document.querySelector('#outlet');
const router = new Router(outlet);
router.setRoutes([
    {path: '/',     component: 'root-view'},
    {path: '/genre/:id',  component: 'genre-view'},
    {path: '/admin/genres',  component: 'genre-table-view'},
    {path: '/admin/genre/:id',  component: 'genre-detail-view'},
    {path: '/serie/:id',  component: 'serie-view'},
    {path: '/admin/series',  component: 'serie-table-view'},
    {path: '/admin/serie/:id',  component: 'serie-detail-view'},
    {path: '/movie/:id',  component: 'movie-view'},
    {path: '/admin/movies',  component: 'movie-table-view'},
    {path: '/admin/movie/:id',  component: 'movie-detail-view'},
    {path: '/search/:query',  component: 'search-view'},
    {path: '/about',  component: 'about-view'},
    {path: '/admin/languages',  component: 'language-table-view'},
    {path: '/admin/language/:id',  component: 'language-detail-view'},
    {path: '/(.*)',  component: 'unknown-view'}
]);
