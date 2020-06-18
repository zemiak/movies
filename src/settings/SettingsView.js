import { html, render } from "lit-html";

export class SettingsView extends HTMLElement {
    connectedCallback() {
        this.render();
    }

    render() {
        render(this.view(), this);
    }

    view() {
        return html`<nav class="panel">
        <p class="panel-heading">
          Settings
        </p>
        <a href="/admin/movies" class="panel-block is-active">
          <span class="panel-icon">
            <i class="fas fa-book" aria-hidden="true"></i>
          </span>
          Movies
        </a>
        <a href="/admin/series" class="panel-block">
          <span class="panel-icon">
            <i class="fas fa-book" aria-hidden="true"></i>
          </span>
          Series
        </a>
        <a href="/admin/genres" class="panel-block">
          <span class="panel-icon">
            <i class="fas fa-book" aria-hidden="true"></i>
          </span>
          Genres
        </a>
        <a href="/admin/languages" class="panel-block">
          <span class="panel-icon">
            <i class="fas fa-book" aria-hidden="true"></i>
          </span>
          Languages
        </a>
      </nav>`;
    }
}

customElements.define("settings-view", SettingsView);
