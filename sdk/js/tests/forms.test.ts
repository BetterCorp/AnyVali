// @vitest-environment jsdom

import { describe, expect, it } from "vitest";
import { object, string, int, array, exportSchema } from "../src/index.js";
import { createFormBindings, initForm } from "../src/forms/index.js";

describe("forms bindings", () => {
  it("derives native field attributes from a schema", () => {
    const schema = object({
      email: string().format("email").minLength(5),
      age: int().min(18),
      tags: array(string()).minItems(1),
    });

    const bindings = createFormBindings({ schema });

    expect(bindings.field("email")).toMatchObject({
      name: "email",
      type: "email",
      required: true,
      minLength: 5,
    });

    expect(bindings.field("age")).toMatchObject({
      name: "age",
      type: "number",
      required: true,
      min: 18,
      step: 1,
    });

    expect(bindings.field("tags")).toMatchObject({
      name: "tags",
      required: true,
    });
  });
});

describe("forms init", () => {
  it("enhances an existing form and blocks invalid submit", () => {
    document.body.innerHTML = `
      <form id="signup">
        <input name="email" />
        <input name="age" />
        <div data-anyvali-error-for="email"></div>
      </form>
    `;

    const schema = exportSchema(
      object({
        email: string().format("email"),
        age: int().min(18),
      })
    );

    const controller = initForm("#signup", {
      schema,
      validateOn: ["blur", "submit"],
    });

    const form = document.querySelector("#signup") as HTMLFormElement;
    const email = form.querySelector('[name="email"]') as HTMLInputElement;
    const age = form.querySelector('[name="age"]') as HTMLInputElement;

    expect(email.type).toBe("email");
    expect(age.type).toBe("number");
    expect(age.getAttribute("min")).toBe("18");

    email.value = "not-an-email";
    age.value = "21";

    const submit = new Event("submit", { cancelable: true });
    const accepted = form.dispatchEvent(submit);

    expect(accepted).toBe(false);
    expect(email.validationMessage.length).toBeGreaterThan(0);
    expect(
      form.querySelector('[data-anyvali-error-for="email"]')?.textContent
    ).not.toBe("");

    controller.destroy();
  });

  it("hooks into htmx validation events for custom validity", () => {
    document.body.innerHTML = `
      <form id="profile">
        <input name="email" />
      </form>
    `;

    const controller = initForm("#profile", {
      schema: exportSchema(object({ email: string().format("email") })),
      htmx: true,
    });

    const input = document.querySelector('[name="email"]') as HTMLInputElement;
    input.value = "broken";
    input.dispatchEvent(new CustomEvent("htmx:validation:validate", { bubbles: true }));

    expect(input.validationMessage.length).toBeGreaterThan(0);

    controller.destroy();
  });

  it("reads nested fields and multi-select values using schema paths", () => {
    document.body.innerHTML = `
      <form id="profile">
        <input name="user.email" />
        <select name="tags" multiple>
          <option value="alpha">Alpha</option>
          <option value="beta">Beta</option>
          <option value="gamma">Gamma</option>
        </select>
        <div data-anyvali-error-for="user.email"></div>
      </form>
    `;

    const controller = initForm("#profile", {
      schema: exportSchema(
        object({
          user: object({
            email: string().format("email"),
          }),
          tags: array(string()).minItems(1),
        })
      ),
      validateOn: ["submit"],
    });

    const form = document.querySelector("#profile") as HTMLFormElement;
    const email = form.querySelector('[name="user.email"]') as HTMLInputElement;
    const tags = form.querySelector('[name="tags"]') as HTMLSelectElement;

    email.value = "team@anyvali.dev";
    tags.options[0].selected = true;
    tags.options[2].selected = true;

    expect(controller.getValues()).toEqual({
      user: { email: "team@anyvali.dev" },
      tags: ["alpha", "gamma"],
    });
    expect(controller.validate()).toBe(true);

    controller.destroy();
  });
});
