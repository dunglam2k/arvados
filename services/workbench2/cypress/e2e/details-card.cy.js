// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

describe('User Details Card tests', function () {
    let activeUser;
    let adminUser;

    before(function () {
        // Only set up common users once. These aren't set up as aliases because
        // aliases are cleaned up after every test. Also it doesn't make sense
        // to set the same users on beforeEach() over and over again, so we
        // separate a little from Cypress' 'Best Practices' here.
        cy.getUser('admin', 'Admin', 'User', true, true)
            .as('adminUser')
            .then(function () {
                adminUser = this.adminUser;
            });
        cy.getUser('activeUser1', 'Active', 'User', false, true)
            .as('activeUser')
            .then(function () {
                activeUser = this.activeUser;
            });
        cy.on('uncaught:exception', (err, runnable) => {
            console.error(err);
        });
    });

    beforeEach(function () {
        cy.clearCookies();
        cy.clearLocalStorage();
    });

    it('should display the user details card', () => {
        cy.loginAs(adminUser);

        cy.get('[data-cy=user-details-card]').should('be.visible');
        cy.get('[data-cy=user-details-card]').contains(adminUser.user.full_name).should('be.visible');
    });

    it('shows the appropriate buttons in the multiselect toolbar', () => {
        const msButtonTooltips = ['View details', 'User account', 'API Details'];

        cy.loginAs(activeUser);

        cy.get('[data-cy=multiselect-button]').should('have.length', msButtonTooltips.length);

        for (let i = 0; i < msButtonTooltips.length; i++) {
            cy.get('[data-cy=multiselect-button]').eq(i).trigger('mouseover');
            cy.get('body').contains(msButtonTooltips[i]).should('exist');
            cy.get('[data-cy=multiselect-button]').eq(i).trigger('mouseout');
        }
    });
});

describe('Project Details Card tests', function () {
    let activeUser;
    let adminUser;

    before(function () {
        // Only set up common users once. These aren't set up as aliases because
        // aliases are cleaned up after every test. Also it doesn't make sense
        // to set the same users on beforeEach() over and over again, so we
        // separate a little from Cypress' 'Best Practices' here.
        cy.getUser('admin', 'Admin', 'User', true, true)
            .as('adminUser')
            .then(function () {
                adminUser = this.adminUser;
            });
        cy.getUser('activeUser1', 'Active', 'User', false, true)
            .as('activeUser')
            .then(function () {
                activeUser = this.activeUser;
            });
        cy.on('uncaught:exception', (err, runnable) => {
            console.error(err);
        });
    });

    beforeEach(function () {
        cy.clearCookies();
        cy.clearLocalStorage();
    });

    it('should display the project details card', () => {
        const projName = `Test project (${Math.floor(999999 * Math.random())})`;
        cy.loginAs(adminUser);

        // Create project
        cy.get('[data-cy=side-panel-button]').click();
        cy.get('[data-cy=side-panel-new-project]').click();
        cy.get('[data-cy=form-dialog]')
            .should('contain', 'New Project')
            .within(() => {
                cy.get('[data-cy=name-field]').within(() => {
                    cy.get('input').type(projName);
                });
            });
        cy.get('[data-cy=form-submit-btn]').click();
        cy.get('[data-cy=form-dialog]').should('not.exist');

        cy.get('[data-cy=project-details-card]').should('be.visible');
        cy.get('[data-cy=project-details-card]').contains(projName).should('be.visible');
    });

    it('shows the appropriate buttons in the multiselect toolbar', () => {
        const msButtonTooltips = ['View details', 'Open in new tab', 'Copy link to clipboard'];

        const msOverflowMenuButtonTooltips = [
            'Open with 3rd party client',
            'API Details',
            'Share',
            'New project',
            'Edit project',
            'Move to',
            'Move to trash',
            'Freeze project',
            'Add to favorites',
        ];

        const projName = `Test project (${Math.floor(999999 * Math.random())})`;
        cy.loginAs(activeUser);

        // Create project
        cy.get('[data-cy=side-panel-button]').click();
        cy.get('[data-cy=side-panel-new-project]').click();
        cy.get('[data-cy=form-dialog]')
            .should('contain', 'New Project')
            .within(() => {
                cy.get('[data-cy=name-field]').within(() => {
                    cy.get('input').type(projName);
                });
            });
        cy.get('[data-cy=form-submit-btn]').click();
        cy.get('[data-cy=form-dialog]').should('not.exist');

        for (let i = 0; i < msButtonTooltips.length; i++) {
            cy.get('[data-cy=multiselect-button]').eq(i).should('exist');
            cy.get('[data-cy=multiselect-button]').eq(i).trigger('mouseover');
            cy.waitForDom()
            cy.get('body').within(() => {
                cy.contains(msButtonTooltips[i]).should('exist');
            });
            cy.get('[data-cy=multiselect-button]').eq(i).trigger('mouseout');
        }

        cy.get('[data-cy=overflow-menu-button]').click();
        cy.get('[data-cy=overflow-menu]')
            .should('be.visible')
            .within(() => {
                cy.get('[data-cy=multiselect-button]').should('exist');

                for (let i = 0; i < msOverflowMenuButtonTooltips.length; i++) {
                    cy.get('li')
                        .eq(i)
                        .within(() => {
                            cy.get(`span`).should('have.prop', 'title', msOverflowMenuButtonTooltips[i]);
                        });
                }
            });
    });

    it('should toggle description display', () => {
        const projName = `Test project (${Math.floor(999999 * Math.random())})`;
        //must be long enough to require a 2nd line
        const projDescription =
            'Science! true daughter of Old Time thou art! Who alterest all things with thy peering eyes. Why preyest thou thus upon the poet’s heart, Vulture, whose wings are dull realities? '
        cy.loginAs(adminUser);

        // Create project
        cy.get('[data-cy=side-panel-button]').click();
        cy.get('[data-cy=side-panel-new-project]').click();
        cy.get('[data-cy=form-dialog]')
            .should('contain', 'New Project')
            .within(() => {
                cy.get('[data-cy=name-field]').within(() => {
                    cy.get('input').type(projName);
                });
            });
        cy.get('[data-cy=form-submit-btn]').click();

        //check for no description
        cy.get('[data-cy=no-description').should('be.visible');

        //add description
        cy.get('[data-cy=side-panel-tree]').contains('Home Projects').click();
        cy.get('[data-cy=project-panel] tbody tr').contains(projName).rightclick({ force: true });
        cy.get('[data-cy=context-menu]').contains('Edit').click();
        cy.get('[data-cy=form-dialog]').within(() => {
            cy.get('div[contenteditable=true]').click().type(projDescription);
            cy.get('[data-cy=form-submit-btn]').click();
        });
        cy.get('[data-cy=project-panel] tbody tr').contains(projName).click({ force: true });
        cy.get('[data-cy=project-details-card]').contains(projName).should('be.visible');

        //toggle description
        //description is always visible, even when collapsed
        cy.get('[data-cy=project-details-card]').contains(projDescription).should('be.visible');
        cy.get('[data-cy=project-details-card]').invoke('height').should('be.lt', 90);
        cy.get('[data-cy=toggle-description]').click();
        cy.waitForDom();
        cy.get('[data-cy=project-details-card]').invoke('height').should('be.gt', 91);
        cy.get('[data-cy=toggle-description]').click();
        cy.waitForDom();
        cy.get('[data-cy=project-details-card]').invoke('height').should('be.lt', 90);
    });

    it('should display key/value pairs', () => {
        const projName = `Test project (${Math.floor(999999 * Math.random())})`;
        cy.loginAs(adminUser);

        // Create project wih key/value pairs
        cy.get('[data-cy=side-panel-button]').click();
        cy.get('[data-cy=side-panel-new-project]').click();
        cy.get('[data-cy=form-dialog]')
            .should('contain', 'New Project')
            .within(() => {
                cy.get('[data-cy=name-field]').within(() => {
                    cy.get('input').type(projName);
                });
            });

        cy.get('[data-cy=key-input]').should('be.visible').click().type('Animal');
        cy.get('[data-cy=value-input]').should('be.visible').click().type('Dog');
        cy.get('[data-cy=property-add-btn]').should('be.visible').click();

        cy.get('[data-cy=key-input]').should('be.visible').click().type('Importance');
        cy.get('[data-cy=value-input]').should('be.visible').click().type('Critical');
        cy.get('[data-cy=property-add-btn]').should('be.visible').click();

        cy.get('[data-cy=form-submit-btn]').click();

        //toggle chips
        cy.get('[data-cy=project-details-card]').invoke('height').should('be.lt', 100);
        cy.get('[data-cy=toggle-chips]').click();
        cy.waitForDom();
        cy.get('[data-cy=project-details-card]').invoke('height').should('be.gt', 101);
        cy.get('[data-cy=toggle-chips').click();
        cy.waitForDom();
        cy.get('[data-cy=project-details-card]').invoke('height').should('be.lt', 100);

        //check for key/value pairs in project details card
        cy.get('[data-cy=project-details-card]').contains('Animal').should('be.visible');
        cy.get('[data-cy=project-details-card]').contains('Importance').should('be.visible').click();
        cy.waitForDom();
        cy.window().then((win) => {
            win.navigator.clipboard.readText().then((text) => {
                //wait is necessary due to known issue with cypress@13.7.1
                cy.wait(1000)
                expect(text).to.match(new RegExp(`Importance: Critical`));
                })
            }
        );
    });
});
