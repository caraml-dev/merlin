import React, { useEffect, useState, useRef } from "react";
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
  const nextPage = () => {
    fetchAPI(limit, cursors[currentPageIdx + 1], search);
    setCurrentPageIdx(currentPageIdx + 1);
  };

  const fetchAPI = (pageLimit, pageCursor, searchQuery) => {
    apiRequest({
      query: {
        limit: pageLimit,
        cursor: pageCursor,
        search: searchQuery
      }
    });
  };

  const prevPage = () => {
    const prevCursor = cursors[currentPageIdx - 1];
    fetchAPI(limit, prevCursor, search);

    setCurrentPageIdx(currentPageIdx - 1);

    // remove cursors
    removeCursor(currentPageIdx);
  };

  const removeCursor = idx => {
    const tempCursors = [...cursors];
    tempCursors.splice(idx);
    setCursors(tempCursors);
  };

  const initialCursors = [""];
  const [cursors, setCursors] = useState(initialCursors);

  useEffect(() => {
    if (search == null) {
      return;
    }
    setCurrentPageIdx(0);
    setCursors(initialCursors);
    fetchAPI(limit, "", search);
  }, [search]);

  const limitHookAlreadyMounted = useRef(false);
  useEffect(() => {
    if (!limitHookAlreadyMounted.current) {
      limitHookAlreadyMounted.current = true;
      return;
    }
    removeCursor(cursors.length - 1);
    fetchAPI(limit, cursors[currentPageIdx], search);
  }, [limit]);

  const nextPageAvailable = () => {
    return cursors[cursors.length - 1] != "";
  };

  const prevPageAvailable = () => {
    return currentPageIdx - 1 >= 0;
  };

  useEffect(() => {
    if (apiResponse.headers == null) {
      return;
    }
    const nextCursor = apiResponse.headers["next-cursor"] || "";
    if (cursors.length > 0 || nextCursor != "") {
      setCursors([...cursors, nextCursor]);
    }
  }, [apiResponse]);

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
