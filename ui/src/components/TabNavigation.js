/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useState } from "react";
import {
  EuiContextMenuItem,
  EuiContextMenuPanel,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiPopover,
  EuiTab,
  EuiTabs
} from "@elastic/eui";

import "./TabNavigation.scss";

const MoreActionsButton = ({ actions }) => {
  const [isPopoverOpen, setPopover] = useState(false);
  const togglePopover = () => setPopover(isPopoverOpen => !isPopoverOpen);

  const items = actions
    .filter(item => !item.hidden)
    .map((item, idx) => (
      <EuiContextMenuItem
        key={idx}
        icon={item.icon}
        onClick={() => {
          togglePopover();
          item.onClick();
        }}
        disabled={item.disabled}
        className={item.color ? `euiTextColor--${item.color}` : ""}>
        {item.name}
      </EuiContextMenuItem>
    ));

  const button = (
    <EuiTabs>
      <EuiTab onClick={togglePopover}>
        <span>
          More Actions&nbsp;
          <EuiIcon
            type="arrowDown"
            size="m"
            style={{ verticalAlign: "text-bottom" }}
          />
        </span>
      </EuiTab>
    </EuiTabs>
  );

  return (
    <EuiPopover
      button={button}
      isOpen={isPopoverOpen}
      closePopover={togglePopover}
      panelPaddingSize="none"
      anchorPosition="downRight">
      <EuiContextMenuPanel
        hasFocus={false}
        className="euiContextPanel--moreActions"
        items={items}
      />
    </EuiPopover>
  );
};

export const TabNavigation = ({ tabs, actions, selectedTab, ...props }) => (
  <EuiFlexGroup direction="row" gutterSize="none">
    <EuiFlexItem grow={true}>
      <EuiTabs>
        {tabs.map((tab, index) => (
          <EuiTab
            {...(tab.href
              ? { href: tab.href, target: tab.target }
              : { onClick: () => props.navigate(`./${tab.id}`) })}
            isSelected={tab.id === selectedTab}
            disabled={tab.disabled}
            key={index}>
            {tab.name}
          </EuiTab>
        ))}
      </EuiTabs>
    </EuiFlexItem>

    {actions && actions.length && (
      <EuiFlexItem grow={false}>
        <MoreActionsButton actions={actions} />
      </EuiFlexItem>
    )}
  </EuiFlexGroup>
);
