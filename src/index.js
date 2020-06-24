import { BreadCrumbs } from "/_dist_/BreadCrumbs.js";
import { RootView } from "/_dist_/root/RootView.js";
import { GenreView } from "/_dist_/genre/GenreView.js";
import { GenreTableView } from "/_dist_/genre/GenreTableView.js";
import { GenreAddView } from "/_dist_/genre/GenreAddView.js";
import { GenreEditView } from "/_dist_/genre/GenreEditView.js";
import { SerieView } from "/_dist_/serie/SerieView.js";
import { SerieTableView } from "/_dist_/serie/SerieTableView.js";
import { SerieAddView } from "/_dist_/serie/SerieAddView.js";
import { SerieEditView } from "/_dist_/serie/SerieEditView.js";
import { MovieView } from "/_dist_/movie/MovieView.js";
import { MovieTableView } from "/_dist_/movie/MovieTableView.js";
import { MovieAddView } from "/_dist_/movie/MovieAddView.js";
import { MovieEditView } from "/_dist_/movie/MovieEditView.js";
import { LanguageAddView } from "/_dist_/language/LanguageAddView.js";
import { LanguageEditView } from "/_dist_/language/LanguageEditView.js";
import { LanguageTableView } from "/_dist_/language/LanguageTableView.js";
import { SearchView } from "/_dist_/search/SearchView.js";
import { UnknownView } from "/_dist_/unknown/UnknownView.js";
import { AboutView } from "/_dist_/about/AboutView.js";
import { ErrorView } from "/_dist_/error/ErrorView.js";
import { SettingsView } from "/_dist_/settings/SettingsView.js";

import { Cache } from "/_dist_/Cache.js";
import { Router } from '@vaadin/router';
import "/_dist_/search/SearchFieldService.js";

new Cache().clear();

const outlet = document.querySelector('#outlet');
const router = new Router(outlet);
router.setRoutes([
    {path: '/',     component: 'root-view'},
    {path: '/genre/:id',  component: 'genre-view'},
    {path: '/admin/genres',  component: 'genre-table-view'},
    {path: '/admin/genre/add',  component: 'genre-add-view'},
    {path: '/admin/genre/:id',  component: 'genre-edit-view'},
    {path: '/serie/:id',  component: 'serie-view'},
    {path: '/admin/series',  component: 'serie-table-view'},
    {path: '/admin/serie/add',  component: 'serie-add-view'},
    {path: '/admin/serie/:id',  component: 'serie-edit-view'},
    {path: '/movie/:id',  component: 'movie-view'},
    {path: '/admin/movies',  component: 'movie-table-view'},
    {path: '/admin/movie/add',  component: 'movie-add-view'},
    {path: '/admin/movie/:id',  component: 'movie-edit-view'},
    {path: '/search/:query',  component: 'search-view'},
    {path: '/about',  component: 'about-view'},
    {path: '/error',  component: 'error-view'},
    {path: '/settings',  component: 'settings-view'},
    {path: '/admin/languages',  component: 'language-table-view'},
    {path: '/admin/language/add',  component: 'language-add-view'},
    {path: '/admin/language/:id',  component: 'language-edit-view'},
    {path: '/(.*)',  component: 'unknown-view'}
]);
