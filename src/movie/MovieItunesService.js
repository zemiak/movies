import { RestClientService } from "../RestClientService.js";

export class MovieItunesService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieItunesService";
    }

    getServicePath() {
        return "/metadata/itunes/" + encodeURI(this.name);
    }

    setName(name) {
        this.name = name;
    }
}
