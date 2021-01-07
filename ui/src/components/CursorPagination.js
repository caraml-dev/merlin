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

import React, { useEffect, useState, useCallback } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiButtonIcon,
  EuiButtonEmpty,
  EuiContextMenuItem,
  EuiContextMenuPanel,
  EuiPopover,
  EuiText
} from "@elastic/eui";

export const CursorPagination = ({
  apiRequest,
  apiResponse,
  search,
  defaultLimit
}) => {
  const [limit, setLimit] = useState(defaultLimit);
  const [currentPageIdx, setCurrentPageIdx] = useState(0);

  const fetchAPI = useCallback(
    (pageLimit, pageCursor, searchQuery) => {
      apiRequest({
        query: {
          limit: pageLimit,
          cursor: pageCursor,
          search: searchQuery
        }
      });
    },
    [apiRequest]
  );

  const removeCursor = (listOfCursor, idx) => {
    const tempCursors = [...listOfCursor];
    tempCursors.splice(idx);
    setCursors(tempCursors);
  };

  const [cursors, setCursors] = useState([""]);

  useEffect(() => {
    if (apiResponse.headers == null) {
      return;
    }
    const nextCursor = apiResponse.headers["next-cursor"] || "";
    setCursors(prev => [...prev, nextCursor]);
  }, [apiResponse, setCursors]);

  const [, setQuery] = useState({ limit, search, cursors });

  useEffect(() => {
    setQuery(prev => {
      const limitChanged = limit !== prev.limit;
      const searchChanged = search !== prev.search;

      var cursorsUsed = cursors;
      if (limitChanged) {
        fetchAPI(limit, cursors[currentPageIdx], search);
        const tempCursors = [...cursors];
        tempCursors.splice(cursors.length - 1);
        setCursors(tempCursors);
      } else if (searchChanged) {
        setCurrentPageIdx(0);
        setCursors([""]);
        fetchAPI(limit, "", search);
      }
      return { limit, search, cursorsUsed };
    });
  }, [limit, search, cursors, currentPageIdx, fetchAPI]);

  const nextPage = () => {
    fetchAPI(limit, cursors[currentPageIdx + 1], search);
    setCurrentPageIdx(currentPageIdx + 1);
  };

  const prevPage = () => {
    const prevCursor = cursors[currentPageIdx - 1];
    fetchAPI(limit, prevCursor, search);
    setCurrentPageIdx(currentPageIdx - 1);

    // remove cursors
    removeCursor(cursors, currentPageIdx);
  };

  const nextPageAvailable = () => {
    return cursors[cursors.length - 1] !== "";
  };

  const prevPageAvailable = () => {
    return currentPageIdx - 1 >= 0;
  };

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  const onButtonClick = () => setIsPopoverOpen(isPopoverOpen => !isPopoverOpen);
  const closePopover = () => setIsPopoverOpen(false);

  const getIconType = size => {
    return size === limit ? "check" : "empty";
  };

  const button = (
    <EuiButtonEmpty
      size="xs"
      color="text"
      iconType="arrowDown"
      iconSide="right"
      onClick={onButtonClick}>
      Rows per page: {limit}
    </EuiButtonEmpty>
  );

  const limitOptions = [20, 50, 100];
  const items = limitOptions.map(value => (
    <EuiContextMenuItem
      key="{value} rows"
      icon={getIconType(value)}
      onClick={() => {
        closePopover();
        setLimit(value);
      }}>
      {value} rows
    </EuiContextMenuItem>
  ));

  const responseLoaded =
    apiResponse && apiResponse.data.length > 0 && apiResponse.isLoaded;
  return responseLoaded ? (
    <EuiFlexGroup justifyContent="spaceBetween" alignItems="center">
      <EuiFlexItem grow={false}>
        <EuiPopover
          button={button}
          isOpen={isPopoverOpen}
          closePopover={closePopover}
          panelPaddingSize="none">
          <EuiContextMenuPanel items={items} />
        </EuiPopover>
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <EuiFlexGroup
          justifyContent="center"
          alignItems="center"
          gutterSize="s">
          <EuiFlexItem grow={false}>
            <EuiButtonIcon
              iconType="arrowLeft"
              disabled={!prevPageAvailable()}
              onClick={prevPage}
              aria-label="Previous Page"
              size="m"
            />
          </EuiFlexItem>
          <EuiFlexItem grow={false}>
            <EuiText size="s">{currentPageIdx + 1}</EuiText>
          </EuiFlexItem>
          <EuiFlexItem grow={false}>
            <EuiButtonIcon
              iconType="arrowRight"
              onClick={nextPage}
              disabled={!nextPageAvailable()}
              aria-label="Next Page"
              size="m"
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
    </EuiFlexGroup>
  ) : null;
};
