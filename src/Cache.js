export class Cache {
    constructor() {
        this.cacheKey = "_com_zemiak_movies_cache";

        const cache = window.localStorage.getItem(this.cacheKey);
        if (! cache) {
            this.clear();
        }
    }

    clear() {
        window.localStorage.setItem(this.cacheKey, JSON.stringify({}));
    }

    get(key) {
        const cache = JSON.parse(window.localStorage.getItem(this.cacheKey));
        if (key in cache) {
            return cache[key];
        }

        return undefined;
    }

    set(key, value) {
        const cache = JSON.parse(window.localStorage.getItem(this.cacheKey));
        cache[key] = value;
        window.localStorage.setItem(this.cacheKey, JSON.stringify(cache));
    }

    contains(key) {
        const cache = JSON.parse(window.localStorage.getItem(this.cacheKey));
        return (key in cache);
    }

    remove(key) {
        const cache = JSON.parse(window.localStorage.getItem(this.cacheKey));
        if (key in cache) {
            delete cache[key];
        }
        window.localStorage.setItem(this.cacheKey, JSON.stringify(cache));
    }
}
