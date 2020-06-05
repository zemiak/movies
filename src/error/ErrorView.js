import { html, render } from "lit-html";

export class ErrorView extends HTMLElement {
    connectedCallback() {
        this.render();
    }

    render() {
        render(this.view(), this);
    }

    view() {
        var errorMessage = window.localStorage.getItem("_com_zemiak_movies_error");

        return html`<section class="hero is-danger is-bold">
        <div class="hero-body">
          <div class="container">
            <h1 class="title">
              Error - Movies
            </h1>
            <p></p>
            <h2 class="subtitle">
                There was an error while reading data from the backend. Please, try again later. <br/>
                ${errorMessage}
            </h2>
          </div>
        </div>
      </section>`;
    }
}

customElements.define("error-view", ErrorView);
